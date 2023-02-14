package lyvecloud

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const resourceIDSeparator = ","

func TestAccS3BucketObjectLockConfiguration_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())
	resourceName := "lyvecloud_s3_bucket_object_lock_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckBucketObjectLockConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectLockConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketObjectLockConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "object_lock_enabled", s3.ObjectLockEnabledEnabled),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.days", "3"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.mode", s3.ObjectLockRetentionModeCompliance),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccS3BucketObjectLockConfiguration_disappears(t *testing.T) {
	rName := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())
	resourceName := "lyvecloud_s3_bucket_object_lock_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckBucketObjectLockConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectLockConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketObjectLockConfigurationExists(resourceName),
					CheckResourceDisappears(testAccProvider, ResourceBucketObjectLockConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccS3BucketObjectLockConfiguration_update(t *testing.T) {
	rName := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())
	resourceName := "lyvecloud_s3_bucket_object_lock_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckBucketObjectLockConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketObjectLockConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketObjectLockConfigurationExists(resourceName),
				),
			},
			{
				Config: testAccBucketObjectLockConfigurationConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "object_lock_enabled", s3.ObjectLockEnabledEnabled),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.years", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.default_retention.0.mode", s3.ObjectLockRetentionModeGovernance),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckBucketObjectLockConfigurationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := testAccProvider.Meta().(Client).S3Client

		bucket, expectedBucketOwner, err := ParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &s3.GetObjectLockConfigurationInput{
			Bucket: aws.String(bucket),
		}

		if expectedBucketOwner != "" {
			input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
		}

		output, err := conn.GetObjectLockConfiguration(input)

		if err != nil {
			return fmt.Errorf("error getting S3 Bucket Object Lock configuration (%s): %w", rs.Primary.ID, err)
		}

		if output == nil || output.ObjectLockConfiguration == nil {
			return fmt.Errorf("S3 Bucket Object Lock configuration (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBucketObjectLockConfigurationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(Client).S3Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "lyvecloud_s3_bucket_object_lock_configuration" {
			continue
		}

		bucket, expectedBucketOwner, err := ParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &s3.GetObjectLockConfigurationInput{
			Bucket: aws.String(bucket),
		}

		if expectedBucketOwner != "" {
			input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
		}

		output, err := conn.GetObjectLockConfiguration(input)

		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket, ErrCodeObjectLockConfigurationNotFound) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting S3 Bucket Object Lock configuration (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("S3 Bucket Object Lock configuration (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

// ParseResourceID is a generic method for parsing an ID string
// for a bucket name and accountID if provided.
func ParseResourceID(id string) (bucket, expectedBucketOwner string, err error) {
	parts := strings.Split(id, resourceIDSeparator)

	if len(parts) == 1 && parts[0] != "" {
		bucket = parts[0]
		return
	}

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		bucket = parts[0]
		expectedBucketOwner = parts[1]
		return
	}

	err = fmt.Errorf("unexpected format for ID (%s), expected BUCKET or BUCKET%sEXPECTED_BUCKET_OWNER", id, resourceIDSeparator)
	return
}

func CheckResourceDisappears(provo *schema.Provider, resource *schema.Resource, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if resourceState.Primary.ID == "" {
			return fmt.Errorf("resource ID missing: %s", resourceName)
		}

		return DeleteResource(resource, resource.Data(resourceState.Primary), provo.Meta())
	}
}

func DeleteResource(resource *schema.Resource, d *schema.ResourceData, meta interface{}) error {
	if resource.DeleteContext != nil || resource.DeleteWithoutTimeout != nil {
		var diags diag.Diagnostics

		if resource.DeleteContext != nil {
			diags = resource.DeleteContext(context.Background(), d, meta)
		} else {
			diags = resource.DeleteWithoutTimeout(context.Background(), d, meta)
		}

		for i := range diags {
			if diags[i].Severity == diag.Error {
				return fmt.Errorf("error deleting resource: %s", diags[i].Summary)
			}
		}

		return nil
	}

	return resource.Delete(d, meta)
}

func testAccBucketObjectLockConfigurationConfig_basic(bucketName string) string {
	return fmt.Sprintf(`
provider "lyvecloud" {
	s3 {}
}

resource "lyvecloud_s3_bucket" "test" {
  bucket = %[1]q

  object_lock_enabled = true
}

resource "lyvecloud_s3_bucket_object_lock_configuration" "test" {
  bucket = lyvecloud_s3_bucket.test.id

  rule {
    default_retention {
      mode = %[2]q
      days = 3
    }
  }
}
`, bucketName, s3.ObjectLockRetentionModeCompliance)
}

func testAccBucketObjectLockConfigurationConfig_update(bucketName string) string {
	return fmt.Sprintf(`
provider "lyvecloud" {
	s3 {}
}

resource "lyvecloud_s3_bucket" "test" {
  bucket = %[1]q

  object_lock_enabled = true
}

resource "lyvecloud_s3_bucket_object_lock_configuration" "test" {
  bucket = lyvecloud_s3_bucket.test.id

  rule {
    default_retention {
      mode  = %[2]q
      years = 1
    }
  }
}
`, bucketName, s3.ObjectLockModeGovernance)
}
