package lyvecloud

import (
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"golang.org/x/net/context"
)

func TestAccLyveCloudS3Bucket_basic(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	region := os.Getenv("LYVECLOUD_REGION")
	resourceName := "lyvecloud_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "region", region),
					resource.TestCheckResourceAttr(resourceName, "bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "object_lock_enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
			},
		},
	})
}

func TestAccS3Bucket_Basic_emptyString(t *testing.T) {
	resourceName := "lyvecloud_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_emptyString,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "bucket", regexp.MustCompile("^terraform-")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func TestAccS3Bucket_Basic_generatedName(t *testing.T) {
	resourceName := "lyvecloud_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "bucket_prefix"},
			},
		},
	})
}

func TestAccS3Bucket_Basic_namePrefix(t *testing.T) {
	resourceName := "lyvecloud_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "bucket", regexp.MustCompile("^tf-test-")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "bucket_prefix"},
			},
		},
	})
}

func TestAccS3Bucket_Basic_forceDestroy(t *testing.T) {
	resourceName := "lyvecloud_s3_bucket.test"
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_forceDestroy(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketAddObjects(resourceName, "data.txt", "prefix/more_data.txt"),
				),
			},
		},
	})
}

func TestAccS3Bucket_Basic_forceDestroyWithEmptyPrefixes(t *testing.T) {
	resourceName := "lyvecloud_s3_bucket.test"
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_forceDestroy(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketAddObjects(resourceName, "data.txt", "/extraleadingslash.txt"),
				),
			},
		},
	})
}

func TestAccS3Bucket_Basic_forceDestroyWithObjectLockEnabled(t *testing.T) {
	resourceName := "lyvecloud_s3_bucket.test"
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_forceDestroyObjectLockEnabledDefaultRetention(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					testAccCheckBucketAddObjects(resourceName, "data.txt", "prefix/more_data.txt"),
				),
			},
		},
	})
}

func TestAccS3Bucket_Tags_basic(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "lyvecloud_s3_bucket.bucket1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_multiTags(rInt),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func TestAccS3Bucket_Tags_withNoSystemTags(t *testing.T) {
	resourceName := "lyvecloud_s3_bucket.test"
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_tags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "AAA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccBucketConfig_updatedTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "XXX"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key4", "DDD"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key5", "EEE"),
				),
			},
			{
				Config: testAccBucketConfig_noTags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			// Verify update from 0 tags.
			{
				Config: testAccBucketConfig_tags(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "AAA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
		},
	})
}

func TestAccS3Bucket_Manage_objectLock(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	resourceName := "lyvecloud_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckBucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_objectLockEnabledNoDefaultRetention(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "object_lock_enabled", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccCheckBucketDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(Client).S3Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "lyvecloud_s3_bucket" {
			continue
		}

		input := &s3.HeadBucketInput{
			Bucket: aws.String(rs.Primary.ID),
		}

		// Retry for S3 eventual consistency
		err := resource.RetryContext(context.Background(), time.Minute, func() *resource.RetryError {
			_, err := conn.HeadBucket(input)

			if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) || tfawserr.ErrCodeEquals(err, "NotFound") {
				return nil
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return resource.RetryableError(fmt.Errorf("Lyve Cloud S3 Bucket still exists: %s", rs.Primary.ID))
		})

		if TimedOut(err) {
			_, err = conn.HeadBucket(input)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func testAccCheckBucketExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		conn := testAccProvider.Meta().(Client).S3Client
		_, err := conn.HeadBucket(&s3.HeadBucketInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
				return fmt.Errorf("S3 bucket not found")
			}
			return err
		}
		return nil
	}
}

func testAccCheckBucketAddObjects(n string, keys ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]
		conn := testAccProvider.Meta().(Client).S3Client

		for _, key := range keys {
			_, err := conn.PutObject(&s3.PutObjectInput{
				Bucket: aws.String(rs.Primary.ID),
				Key:    aws.String(key),
			})

			if err != nil {
				return fmt.Errorf("PutObject error: %s", err)
			}
		}

		return nil
	}
}

func testAccBucketConfig_basic(randInt string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "test" {
  bucket = "%s"
}
`, randInt)
}

func testAccBucketConfig_forceDestroy(bucketName string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "test" {
  bucket        = "%s"
  force_destroy = true
}
`, bucketName)
}

func testAccBucketConfig_forceDestroyObjectLockEnabledDefaultRetention(bucketName string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "test" {
  bucket        = "%s"
  force_destroy = true

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

`, bucketName, s3.ObjectLockRetentionModeGovernance)
}

func testAccBucketConfig_multiTags(randInt int) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "bucket1" {
  bucket        = "tf-test-bucket-1-%[1]d"
  force_destroy = true

  tags = {
    Name        = "tf-test-bucket-1-%[1]d"
    Environment = "%[1]d"
  }
}

resource "lyvecloud_s3_bucket" "bucket2" {
  bucket        = "tf-test-bucket-2-%[1]d"
  force_destroy = true

  tags = {
    Name        = "tf-test-bucket-2-%[1]d"
    Environment = "%[1]d"
  }
}

resource "lyvecloud_s3_bucket" "bucket3" {
  bucket        = "tf-test-bucket-3-%[1]d"
  force_destroy = true

  tags = {
    Name        = "tf-test-bucket-3-%[1]d"
    Environment = "%[1]d"
  }
}

resource "lyvecloud_s3_bucket" "bucket4" {
  bucket        = "tf-test-bucket-4-%[1]d"
  force_destroy = true

  tags = {
    Name        = "tf-test-bucket-4-%[1]d"
    Environment = "%[1]d"
  }
}

resource "lyvecloud_s3_bucket" "bucket5" {
  bucket        = "tf-test-bucket-5-%[1]d"
  force_destroy = true

  tags = {
    Name        = "tf-test-bucket-5-%[1]d"
    Environment = "%[1]d"
  }
}

resource "lyvecloud_s3_bucket" "bucket6" {
  bucket        = "tf-test-bucket-6-%[1]d"
  force_destroy = true

  tags = {
    Name        = "tf-test-bucket-6-%[1]d"
    Environment = "%[1]d"
  }
}
`, randInt)
}

func testAccBucketConfig_tags(bucketName string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = false

  tags = {
    Key1 = "AAA"
    Key2 = "BBB"
    Key3 = "CCC"
  }
}
`, bucketName)
}

func testAccBucketConfig_updatedTags(bucketName string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = false

  tags = {
    Key2 = "BBB"
    Key3 = "XXX"
    Key4 = "DDD"
    Key5 = "EEE"
  }
}
`, bucketName)
}

func testAccBucketConfig_noTags(bucketName string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = false
}
`, bucketName)
}

func testAccBucketConfig_objectLockEnabledNoDefaultRetention(bucketName string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "test" {
  bucket = %[1]q

  object_lock_enabled = true
}
`, bucketName)
}

const testAccBucketConfig_emptyString = `
resource "lyvecloud_s3_bucket" "test" {
  bucket = ""
}
`

const testAccBucketConfig_generatedName = `
resource "lyvecloud_s3_bucket" "test" {
  bucket_prefix = "tf-test-"
}
`

const testAccBucketConfig_namePrefix = `
resource "lyvecloud_s3_bucket" "test" {
  bucket_prefix = "tf-test-"
}
`
