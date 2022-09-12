package lyvecloud

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceObjectCopy() *schema.Resource {
	return &schema.Resource{
		Create: resourceObjectCopyCreate,
		Read:   resourceObjectCopyRead,
		Update: resourceObjectCopyUpdate,
		Delete: resourceObjectCopyDelete,

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"cache_control": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"content_disposition": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"content_encoding": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"content_language": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"content_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"copy_if_match": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"copy_if_modified_since": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"copy_if_none_match": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"copy_if_unmodified_since": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"last_modified": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"metadata": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"metadata_directive": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(s3.MetadataDirective_Values(), false),
			},
			"source": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"source_version_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tagging_directive": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(s3.TaggingDirective_Values(), false),
			},
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"version_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceObjectCopyCreate(d *schema.ResourceData, meta interface{}) error {
	return resourceObjectCopyDoCopy(d, meta)
}

func resourceObjectCopyRead(d *schema.ResourceData, meta interface{}) error {
	conn := *meta.(Client).S3Client

	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)

	resp, err := conn.HeadObject(
		&s3.HeadObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})

	if !d.IsNewResource() && tfawserr.ErrStatusCodeEquals(err, http.StatusNotFound) {
		log.Printf("[WARN] S3 Object (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading S3 Object (%s): %w", d.Id(), err)
	}

	if resp == nil {
		return fmt.Errorf("error reading S3 Object (%s): empty response", d.Id())
	}

	log.Printf("[DEBUG] Reading S3 Object meta: %s", resp)

	d.Set("cache_control", resp.CacheControl)
	d.Set("content_disposition", resp.ContentDisposition)
	d.Set("content_encoding", resp.ContentEncoding)
	d.Set("content_language", resp.ContentLanguage)
	d.Set("content_type", resp.ContentType)
	metadata := PointersMapToStringList(resp.Metadata)

	// AWS Go SDK capitalizes metadata, this is a workaround. https://github.com/aws/aws-sdk-go/issues/445
	for k, v := range metadata {
		delete(metadata, k)
		metadata[strings.ToLower(k)] = v
	}

	if err := d.Set("metadata", metadata); err != nil {
		return fmt.Errorf("error setting metadata: %w", err)
	}

	d.Set("version_id", resp.VersionId)

	// See https://forums.aws.amazon.com/thread.jspa?threadID=44003
	d.Set("etag", strings.Trim(aws.StringValue(resp.ETag), `"`))

	// Retry due to S3 eventual consistency
	tagsRaw, err := RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return ObjectListTags(&conn, bucket, key)
	}, s3.ErrCodeNoSuchBucket)

	if err != nil {
		return fmt.Errorf("error listing tags for S3 Bucket (%s) Object (%s): %w", bucket, key, err)
	}

	tags, ok := tagsRaw.(KeyValueTags)

	if !ok {
		return fmt.Errorf("error listing tags for S3 Bucket (%s) Object (%s): unable to convert tags", bucket, key)
	}

	if err := d.Set("tags", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceObjectCopyUpdate(d *schema.ResourceData, meta interface{}) error {
	// if any of these exist, let the API decide whether to copy
	for _, key := range []string{
		"copy_if_match",
		"copy_if_modified_since",
		"copy_if_none_match",
		"copy_if_unmodified_since",
	} {
		if _, ok := d.GetOk(key); ok {
			return resourceObjectCopyDoCopy(d, meta)
		}
	}

	args := []string{
		"bucket",
		"cache_control",
		"content_disposition",
		"content_encoding",
		"content_language",
		"content_type",
		"key",
		"metadata",
		"metadata_directive",
		"source",
		"tagging_directive",
		"tags",
	}
	if d.HasChanges(args...) {
		return resourceObjectCopyDoCopy(d, meta)
	}

	return nil
}

func resourceObjectCopyDelete(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(S3, meta.(Client)) {
		return fmt.Errorf("credentials for S3 operations are missing")
	}

	conn := *meta.(Client).S3Client

	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)
	// We are effectively ignoring all leading '/'s in the key name and
	// treating multiple '/'s as a single '/' as aws.Config.DisableRestProtocolURICleaning is false
	key = strings.TrimLeft(key, "/")
	key = regexp.MustCompile(`/+`).ReplaceAllString(key, "/")

	err := deleteObjectVersion(&conn, bucket, key, "", false)

	if err != nil {
		return fmt.Errorf("error deleting S3 Bucket (%s) Object (%s): %w", bucket, key, err)
	}
	return nil
}

func resourceObjectCopyDoCopy(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(S3, meta.(Client)) {
		return fmt.Errorf("credentials for S3 operations are missing")
	}

	conn := *meta.(Client).S3Client
	tags := New(d.Get("tags").(map[string]interface{}))

	input := &s3.CopyObjectInput{
		Bucket:     aws.String(d.Get("bucket").(string)),
		Key:        aws.String(d.Get("key").(string)),
		CopySource: aws.String(url.QueryEscape(d.Get("source").(string))),
	}

	if v, ok := d.GetOk("cache_control"); ok {
		input.CacheControl = aws.String(v.(string))
	}

	if v, ok := d.GetOk("content_disposition"); ok {
		input.ContentDisposition = aws.String(v.(string))
	}

	if v, ok := d.GetOk("content_encoding"); ok {
		input.ContentEncoding = aws.String(v.(string))
	}

	if v, ok := d.GetOk("content_language"); ok {
		input.ContentLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("content_type"); ok {
		input.ContentType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("copy_if_match"); ok {
		input.CopySourceIfMatch = aws.String(v.(string))
	}

	if v, ok := d.GetOk("copy_if_modified_since"); ok {
		input.CopySourceIfModifiedSince = expandObjectDate(v.(string))
	}

	if v, ok := d.GetOk("copy_if_none_match"); ok {
		input.CopySourceIfNoneMatch = aws.String(v.(string))
	}

	if v, ok := d.GetOk("copy_if_unmodified_since"); ok {
		input.CopySourceIfUnmodifiedSince = expandObjectDate(v.(string))
	}

	if v, ok := d.GetOk("metadata"); ok {
		input.Metadata = ExpandStringMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("metadata_directive"); ok {
		input.MetadataDirective = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tagging_directive"); ok {
		input.TaggingDirective = aws.String(v.(string))
	}

	if len(tags) > 0 {
		// The tag-set must be encoded as URL Query parameters.
		input.Tagging = aws.String(tags.URLEncode())
	}

	output, err := conn.CopyObject(input)
	if err != nil {
		return fmt.Errorf("error copying S3 object (bucket: %s; key: %s; source: %s): %w", aws.StringValue(input.Bucket), aws.StringValue(input.Key), aws.StringValue(input.CopySource), err)
	}

	if output.CopyObjectResult != nil {
		d.Set("etag", strings.Trim(aws.StringValue(output.CopyObjectResult.ETag), `"`))
		d.Set("last_modified", flattenObjectDate(output.CopyObjectResult.LastModified))
	}

	d.Set("source_version_id", output.CopySourceVersionId)
	d.Set("version_id", output.VersionId)

	d.SetId(d.Get("key").(string))
	return resourceObjectRead(d, meta)
}

func expandObjectDate(v string) *time.Time {
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return nil
	}

	return aws.Time(t)
}
