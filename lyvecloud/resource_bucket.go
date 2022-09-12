package lyvecloud

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	bucketCreatedTimeout = 2 * time.Minute
	propagationTimeout   = 1 * time.Minute
)

func ResourceBucket() *schema.Resource {
	return &schema.Resource{
		Create:        resourceBucketCreate,
		Read:          resourceBucketRead,
		Update:        resourceBucketUpdate,
		DeleteContext: resourceBucketDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 63),
			},
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"object_lock_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceBucketCreate(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(S3, meta.(Client)) {
		return fmt.Errorf("credentials for S3 operations are missing")
	}

	conn := *meta.(Client).S3Client
	bucket := d.Get("bucket").(string)

	req := &s3.CreateBucketInput{
		Bucket:                     aws.String(bucket),
		ObjectLockEnabledForBucket: aws.Bool(d.Get("object_lock_enabled").(bool)),
	}

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.CreateBucket(req)

		if tfawserr.ErrCodeEquals(err, ErrCodeOperationAborted) {
			return resource.RetryableError(fmt.Errorf("error creating S3 Bucket (%s), retrying: %w", bucket, err))
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if TimedOut(err) {
		_, err = conn.CreateBucket(req)
	}
	if err != nil {
		return fmt.Errorf("error creating S3 Bucket (%s): %w", bucket, err)
	}

	// Assign the bucket name as the resource ID
	d.SetId(bucket)
	return resourceBucketUpdate(d, meta)
}

func resourceBucketUpdate(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(S3, meta.(Client)) {
		return fmt.Errorf("credentials for S3 operations are missing")
	}

	conn := *meta.(Client).S3Client

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		// Retry due to S3 eventual consistency
		_, err := RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
			terr := BucketUpdateTags(&conn, d.Id(), o, n)
			return nil, terr
		}, s3.ErrCodeNoSuchBucket)
		if err != nil {
			return fmt.Errorf("error updating S3 Bucket (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceBucketRead(d, meta)
}

func resourceBucketRead(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(S3, meta.(Client)) {
		return fmt.Errorf("credentials for S3 operations are missing")
	}

	conn := *meta.(Client).S3Client

	input := &s3.HeadBucketInput{
		Bucket: aws.String(d.Id()),
	}

	err := resource.Retry(bucketCreatedTimeout, func() *resource.RetryError {
		_, err := conn.HeadBucket(input)

		if d.IsNewResource() && tfawserr.ErrStatusCodeEquals(err, http.StatusNotFound) {
			return resource.RetryableError(err)
		}

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if TimedOut(err) {
		_, err = conn.HeadBucket(input)
	}

	if !d.IsNewResource() && tfawserr.ErrStatusCodeEquals(err, http.StatusNotFound) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading S3 Bucket (%s): %w", d.Id(), err)
	}

	// Add the region as an attribute
	discoveredRegion, err := RetryWhenAWSErrCodeEquals(d.Timeout(schema.TimeoutRead), func() (interface{}, error) {
		return s3manager.GetBucketRegionWithClient(context.Background(), &conn, d.Id(), func(r *request.Request) {
			// By default, GetBucketRegion forces virtual host addressing, which
			// is not compatible with many non-AWS implementations. Instead, pass
			// the provider s3_force_path_style configuration, which defaults to
			// false, but allows override.
			r.Config.S3ForcePathStyle = conn.Config.S3ForcePathStyle

			// By default, GetBucketRegion uses anonymous credentials when doing
			// a HEAD request to get the bucket region. This breaks in aws-cn regions
			// when the account doesn't have an ICP license to host public content.
			// Use the current credentials when getting the bucket region.
			r.Config.Credentials = conn.Config.Credentials
		})
	}, "NotFound")

	// The S3 API method calls above can occasionally return no error (i.e. NoSuchBucket)
	// after a bucket has been deleted (eventual consistency woes :/), thus, when making extra S3 API calls
	// such as s3manager.GetBucketRegionWithClient, the error should be caught for non-new buckets as follows.
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting S3 Bucket location: %s", err)
	}

	region := discoveredRegion.(string)
	if err := d.Set("region", region); err != nil {
		return err
	}

	d.Set("bucket", d.Id())

	return nil
}

func resourceBucketDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if CheckCredentials(S3, meta.(Client)) {
		return diag.FromErr(fmt.Errorf("credentials for S3 operations are missing"))
	}

	conn := *meta.(Client).S3Client

	_, err := conn.DeleteBucketWithContext(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		return nil
	}

	if tfawserr.ErrCodeEquals(err, ErrCodeBucketNotEmpty) {
		if d.Get("force_destroy").(bool) {

			// bucket may have things delete them
			log.Printf("[DEBUG] S3 Bucket attempting to forceDestroy %s", err)

			// Delete everything including locked objects.
			// Don't ignore any object errors or we could recurse infinitely.
			objectLockEnabled := d.Get("object_lock_enabled").(bool)

			if n, err := EmptyBucket(ctx, &conn, d.Id(), objectLockEnabled); err != nil {
				return diag.Errorf("emptying S3 Bucket (%s): %s", d.Id(), err)
			} else {
				log.Printf("[DEBUG] Deleted %d S3 objects", n)
			}

			// this line recurses until all objects are deleted or an error is returned
			return resourceBucketDelete(ctx, d, meta)
		}
	}

	if err != nil {
		return diag.Errorf("deleting S3 Bucket (%s): %s", d.Id(), err)
	}

	return nil
}
