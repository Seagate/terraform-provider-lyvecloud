package lyvecloud

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceBucket() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceBucketRead,

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"region": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceBucketRead(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(S3, meta.(Client)) {
		return fmt.Errorf("credentials for S3 operations are missing")
	}

	conn := *meta.(Client).S3Client

	bucket := d.Get("bucket").(string)

	input := &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	}

	log.Printf("[DEBUG] Reading S3 bucket: %s", input)

	_, err := conn.HeadBucket(input)
	if err != nil {
		return fmt.Errorf("failed getting S3 bucket (%s): %w", bucket, err)
	}

	d.SetId(bucket)

	err = bucketLocation(meta.(Client).S3Client, d, bucket)
	if err != nil {
		return fmt.Errorf("error getting S3 Bucket location: %w", err)
	}

	return nil
}

func bucketLocation(client *s3.S3, d *schema.ResourceData, bucket string) error {
	region, err := s3manager.GetBucketRegionWithClient(context.Background(), client, bucket, func(r *request.Request) {
		r.Config.S3ForcePathStyle = client.Config.S3ForcePathStyle
		r.Config.Credentials = client.Config.Credentials
	})
	if err != nil {
		return err
	}

	if err := d.Set("region", region); err != nil {
		return err
	}
	return nil
}
