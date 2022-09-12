package lyvecloud

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/mitchellh/go-homedir"
)

const objectCreationTimeout = 2 * time.Minute

type ResourceDiffer interface {
	HasChange(string) bool
}

func ResourceObject() *schema.Resource {
	return &schema.Resource{
		Create: resourceObjectCreate,
		Read:   resourceObjectRead,
		Update: resourceObjectUpdate,
		Delete: resourceObjectDelete,

		Importer: &schema.ResourceImporter{
			State: resourceObjectImport,
		},

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
			},
			"content_disposition": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"content_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"metadata": {
				Type:         schema.TypeMap,
				Optional:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
				ValidateFunc: validateMetadataIsLowerCase,
			},
			"key": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"source": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
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

func resourceObjectCreate(d *schema.ResourceData, meta interface{}) error {
	return resourceObjectUpload(d, meta)
}

func resourceObjectRead(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(S3, meta.(Client)) {
		return fmt.Errorf("credentials for S3 operations are missing")
	}

	conn := meta.(Client).S3Client

	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)

	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	var resp *s3.HeadObjectOutput

	err := resource.Retry(objectCreationTimeout, func() *resource.RetryError {
		var err error

		resp, err = conn.HeadObject(input)

		if d.IsNewResource() && tfawserr.ErrStatusCodeEquals(err, http.StatusNotFound) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if TimedOut(err) {
		resp, err = conn.HeadObject(input)
	}

	if !d.IsNewResource() && tfawserr.ErrStatusCodeEquals(err, http.StatusNotFound) {
		log.Printf("[WARN] S3 Object (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading S3 Object (%s): %w", d.Id(), err)
	}

	d.Set("cache_control", resp.CacheControl)
	d.Set("content_disposition", resp.ContentDisposition)
	d.Set("content_encoding", resp.ContentEncoding)
	d.Set("content_type", resp.ContentType)
	d.Set("etag", strings.Trim(aws.StringValue(resp.ETag), `"`))
	d.Set("version_id", resp.VersionId)

	metadata := PointersMapToStringList(resp.Metadata)

	// AWS Go SDK capitalizes metadata, this is a workaround. https://github.com/aws/aws-sdk-go/issues/445
	for k, v := range metadata {
		delete(metadata, k)
		metadata[strings.ToLower(k)] = v
	}

	if err := d.Set("metadata", metadata); err != nil {
		return fmt.Errorf("error setting metadata: %s", err)
	}

	// Retry due to S3 eventual consistency
	tagsRaw, err := RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return ObjectListTags(conn, bucket, key)
	}, s3.ErrCodeNoSuchBucket)

	if err != nil {
		return fmt.Errorf("error listing tags for S3 Bucket (%s) Object (%s): %s", bucket, key, err)
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

func resourceObjectUpdate(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(S3, meta.(Client)) {
		return fmt.Errorf("credentials for S3 operations are missing")
	}

	if hasObjectContentChanges(d) {
		return resourceObjectUpload(d, meta)
	}

	conn := meta.(Client).S3Client

	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := ObjectUpdateTags(conn, bucket, key, o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceObjectRead(d, meta)
}

func resourceObjectDelete(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(S3, meta.(Client)) {
		return fmt.Errorf("credentials for S3 operations are missing")
	}

	conn := meta.(Client).S3Client

	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)
	// We are effectively ignoring all leading '/'s in the key name and
	// treating multiple '/'s as a single '/' as aws.Config.DisableRestProtocolURICleaning is false
	key = strings.TrimLeft(key, "/")
	key = regexp.MustCompile(`/+`).ReplaceAllString(key, "/")

	var err error
	if _, ok := d.GetOk("version_id"); ok {
		_, err = DeleteAllObjectVersions(conn, bucket, key, d.Get("force_destroy").(bool), false)
	} else {
		err = deleteObjectVersion(conn, bucket, key, "", false)
	}

	if err != nil {
		return fmt.Errorf("error deleting S3 Bucket (%s) Object (%s): %s", bucket, key, err)
	}

	return nil
}

func resourceObjectImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	id := d.Id()
	id = strings.TrimPrefix(id, "s3://")
	parts := strings.Split(id, "/")

	if len(parts) < 2 {
		return []*schema.ResourceData{d}, fmt.Errorf("id %s should be in format <bucket>/<key> or s3://<bucket>/<key>", id)
	}

	bucket := parts[0]
	key := strings.Join(parts[1:], "/")

	d.SetId(key)
	d.Set("bucket", bucket)
	d.Set("key", key)

	return []*schema.ResourceData{d}, nil
}

func resourceObjectUpload(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(S3, meta.(Client)) {
		return fmt.Errorf("credentials for S3 operations are missing")
	}

	conn := meta.(Client).S3Client
	uploader := s3manager.NewUploaderWithClient(conn)
	tags := New(d.Get("tags").(map[string]interface{}))

	var body io.ReadSeeker

	if v, ok := d.GetOk("source"); ok {
		source := v.(string)
		path, err := homedir.Expand(source)
		if err != nil {
			return fmt.Errorf("error expanding homedir in source (%s): %s", source, err)
		}
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("error opening S3 object source (%s): %s", path, err)
		}

		body = file
		defer func() {
			err := file.Close()
			if err != nil {
				log.Printf("[WARN] Error closing S3 object source (%s): %s", path, err)
			}
		}()
	} else {
		body = bytes.NewReader([]byte{})
	}

	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)

	input := &s3manager.UploadInput{
		Body:   body,
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
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

	if v, ok := d.GetOk("content_type"); ok {
		input.ContentType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("metadata"); ok {
		input.Metadata = ExpandStringMap(v.(map[string]interface{}))
	}

	if len(tags) > 0 {
		// The tag-set must be encoded as URL Query parameters.
		input.Tagging = aws.String(tags.URLEncode())
	}

	if _, err := uploader.Upload(input); err != nil {
		return fmt.Errorf("error uploading object to S3 bucket (%s): %s", bucket, err)
	}

	d.SetId(key)

	return resourceObjectRead(d, meta)
}

// DeleteAllObjectVersions deletes all versions of a specified key from an S3 bucket.
// If key is empty then all versions of all objects are deleted.
// Set force to true to override any S3 object lock protections on object lock enabled buckets.
// Returns the number of objects deleted.
func DeleteAllObjectVersions(conn *s3.S3, bucketName, key string, force, ignoreObjectErrors bool) (int64, error) {
	var nObjects int64

	input := &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucketName),
	}
	if key != "" {
		input.Prefix = aws.String(key)
	}

	var lastErr error
	err := conn.ListObjectVersionsPages(input, func(page *s3.ListObjectVersionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, objectVersion := range page.Versions {
			objectKey := aws.StringValue(objectVersion.Key)
			objectVersionID := aws.StringValue(objectVersion.VersionId)

			if key != "" && key != objectKey {
				continue
			}

			err := deleteObjectVersion(conn, bucketName, objectKey, objectVersionID, force)

			if err == nil {
				nObjects++
			}

			if tfawserr.ErrCodeEquals(err, "AccessDenied") && force {
				// Remove any legal hold.
				resp, err := conn.HeadObject(&s3.HeadObjectInput{
					Bucket:    aws.String(bucketName),
					Key:       objectVersion.Key,
					VersionId: objectVersion.VersionId,
				})

				if err != nil {
					log.Printf("[ERROR] Error getting S3 Bucket (%s) Object (%s) Version (%s) metadata: %s", bucketName, objectKey, objectVersionID, err)
					lastErr = err
					continue
				}

				if aws.StringValue(resp.ObjectLockLegalHoldStatus) == s3.ObjectLockLegalHoldStatusOn {
					_, err := conn.PutObjectLegalHold(&s3.PutObjectLegalHoldInput{
						Bucket:    aws.String(bucketName),
						Key:       objectVersion.Key,
						VersionId: objectVersion.VersionId,
						LegalHold: &s3.ObjectLockLegalHold{
							Status: aws.String(s3.ObjectLockLegalHoldStatusOff),
						},
					})

					if err != nil {
						log.Printf("[ERROR] Error putting S3 Bucket (%s) Object (%s) Version(%s) legal hold: %s", bucketName, objectKey, objectVersionID, err)
						lastErr = err
						continue
					}

					// Attempt to delete again.
					err = deleteObjectVersion(conn, bucketName, objectKey, objectVersionID, force)

					if err != nil {
						lastErr = err
					} else {
						nObjects++
					}

					continue
				}

				// AccessDenied for another reason.
				lastErr = fmt.Errorf("AccessDenied deleting S3 Bucket (%s) Object (%s) Version: %s", bucketName, objectKey, objectVersionID)
				continue
			}

			if err != nil {
				lastErr = err
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		err = nil
	}

	if err != nil {
		return nObjects, err
	}

	if lastErr != nil {
		if !ignoreObjectErrors {
			return nObjects, fmt.Errorf("error deleting at least one object version, last error: %s", lastErr)
		}

		lastErr = nil
	}

	err = conn.ListObjectVersionsPages(input, func(page *s3.ListObjectVersionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, deleteMarker := range page.DeleteMarkers {
			deleteMarkerKey := aws.StringValue(deleteMarker.Key)
			deleteMarkerVersionID := aws.StringValue(deleteMarker.VersionId)

			if key != "" && key != deleteMarkerKey {
				continue
			}

			// Delete markers have no object lock protections.
			err := deleteObjectVersion(conn, bucketName, deleteMarkerKey, deleteMarkerVersionID, false)

			if err != nil {
				lastErr = err
			} else {
				nObjects++
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		err = nil
	}

	if err != nil {
		return nObjects, err
	}

	if lastErr != nil {
		if !ignoreObjectErrors {
			return nObjects, fmt.Errorf("error deleting at least one object delete marker, last error: %s", lastErr)
		}

		lastErr = nil
	}

	return nObjects, nil
}

// deleteObjectVersion deletes a specific object version.
// Set force to true to override any S3 object lock protections.
func deleteObjectVersion(conn *s3.S3, b, k, v string, force bool) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(b),
		Key:    aws.String(k),
	}

	if v != "" {
		input.VersionId = aws.String(v)
	}

	if force {
		input.BypassGovernanceRetention = aws.Bool(true)
	}

	log.Printf("[INFO] Deleting S3 Bucket (%s) Object (%s) Version: %s", b, k, v)
	_, err := conn.DeleteObject(input)

	if err != nil {
		log.Printf("[WARN] Error deleting S3 Bucket (%s) Object (%s) Version (%s): %s", b, k, v, err)
	}

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) || tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchKey) {
		return nil
	}

	return err
}

func flattenObjectDate(t *time.Time) string {
	if t == nil {
		return ""
	}

	return t.Format(time.RFC3339)
}

func hasObjectContentChanges(d ResourceDiffer) bool {
	for _, key := range []string{
		"cache_control",
		"content_disposition",
		"content_encoding",
		"content_language",
		"content_type",
		"metadata",
		"source",
	} {
		if d.HasChange(key) {
			return true
		}
	}
	return false
}

func validateMetadataIsLowerCase(v interface{}, k string) (ws []string, errors []error) {
	value := v.(map[string]interface{})

	for k := range value {
		if k != strings.ToLower(k) {
			errors = append(errors, fmt.Errorf(
				"metadata must be lowercase only. Offending key: %q", k))
		}
	}
	return
}
