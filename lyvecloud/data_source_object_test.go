package lyvecloud

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const rfc1123RegexPattern = `^[a-zA-Z]{3}, [0-9]+ [a-zA-Z]+ [0-9]{4} [0-9:]+ [A-Z]+$`

func TestAccS3ObjectDataSource_basic(t *testing.T) {
	rInt := sdkacctest.RandInt()

	var rObj s3.GetObjectOutput
	var dsObj s3.GetObjectOutput

	resourceName := "lyvecloud_s3_object.object"
	dataSourceName := "data.lyvecloud_s3_object.obj"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		ProviderFactories:         testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectDataSourceConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &rObj),
					testAccCheckObjectExistsDataSource(dataSourceName, &dsObj),
					resource.TestCheckResourceAttr(dataSourceName, "content_length", "11"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_type", resourceName, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName, "last_modified", regexp.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_mode", resourceName, "object_lock_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_retain_until_date", resourceName, "object_lock_retain_until_date"),
					resource.TestCheckNoResourceAttr(dataSourceName, "body"),
				),
			},
		},
	})
}

func TestAccS3ObjectDataSource_readableBody(t *testing.T) {
	rInt := sdkacctest.RandInt()

	var rObj s3.GetObjectOutput
	var dsObj s3.GetObjectOutput

	resourceName := "lyvecloud_s3_object.object"
	dataSourceName := "data.lyvecloud_s3_object.obj"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		ProviderFactories:         testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectDataSourceConfig_readableBody(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &rObj),
					testAccCheckObjectExistsDataSource(dataSourceName, &dsObj),
					resource.TestCheckResourceAttr(dataSourceName, "content_length", "3"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_type", resourceName, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName, "last_modified", regexp.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_mode", resourceName, "object_lock_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_retain_until_date", resourceName, "object_lock_retain_until_date"),
					resource.TestCheckResourceAttr(dataSourceName, "body", "yes"),
				),
			},
		},
	})
}

func TestAccS3ObjectDataSource_allParams(t *testing.T) {
	rInt := sdkacctest.RandInt()

	var rObj s3.GetObjectOutput
	var dsObj s3.GetObjectOutput

	resourceName := "lyvecloud_s3_object.object"
	dataSourceName := "data.lyvecloud_s3_object.obj"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		ProviderFactories:         testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectDataSourceConfig_allParams(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &rObj),
					testAccCheckObjectExistsDataSource(dataSourceName, &dsObj),
					resource.TestCheckResourceAttr(dataSourceName, "content_length", "25"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_type", resourceName, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName, "last_modified", regexp.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttrPair(dataSourceName, "version_id", resourceName, "version_id"),
					resource.TestCheckNoResourceAttr(dataSourceName, "body"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cache_control", resourceName, "cache_control"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_disposition", resourceName, "content_disposition"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_encoding", resourceName, "content_encoding"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_language", resourceName, "content_language"),
					// Encryption is off
					resource.TestCheckResourceAttrPair(dataSourceName, "server_side_encryption", resourceName, "server_side_encryption"),
					resource.TestCheckResourceAttr(dataSourceName, "metadata.%", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_mode", resourceName, "object_lock_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_retain_until_date", resourceName, "object_lock_retain_until_date"),
				),
			},
		},
	})
}

func testAccCheckObjectExistsDataSource(n string, obj *s3.GetObjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find S3 object data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("S3 object data source ID not set")
		}

		conn := testAccProvider.Meta().(Client).S3Client
		out, err := conn.GetObject(
			&s3.GetObjectInput{
				Bucket: aws.String(rs.Primary.Attributes["bucket"]),
				Key:    aws.String(rs.Primary.Attributes["key"]),
			})
		if err != nil {
			return fmt.Errorf("Failed getting S3 Object from %s: %s",
				rs.Primary.Attributes["bucket"]+"/"+rs.Primary.Attributes["key"], err)
		}

		*obj = *out

		return nil
	}
}

func testAccObjectDataSourceConfig_basic(randInt int) string {
	return fmt.Sprintf(`
provider "lyvecloud" {
	s3 {}
}

resource "lyvecloud_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%[1]d"
}

resource "lyvecloud_s3_object" "object" {
  bucket  = lyvecloud_s3_bucket.object_bucket.bucket
  key     = "tf-testing-obj-%[1]d"
  content = "Hello World"
}

data "lyvecloud_s3_object" "obj" {
  bucket = lyvecloud_s3_bucket.object_bucket.bucket
  key    = lyvecloud_s3_object.object.key
}
`, randInt)
}

func testAccObjectDataSourceConfig_readableBody(randInt int) string {
	return fmt.Sprintf(`
provider "lyvecloud" {
	s3 {}
}

resource "lyvecloud_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%[1]d"
}

resource "lyvecloud_s3_object" "object" {
  bucket       = lyvecloud_s3_bucket.object_bucket.bucket
  key          = "tf-testing-obj-%[1]d-readable"
  content      = "yes"
  content_type = "text/plain"
}

data "lyvecloud_s3_object" "obj" {
  bucket = lyvecloud_s3_bucket.object_bucket.bucket
  key    = lyvecloud_s3_object.object.key
}
`, randInt)
}

func testAccObjectDataSourceConfig_allParams(randInt int) string {
	return fmt.Sprintf(`
provider "lyvecloud" {
	s3 {}
}

resource "lyvecloud_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%[1]d"
  object_lock_enabled = true
}

resource "lyvecloud_s3_object" "object" {
  bucket = lyvecloud_s3_bucket.object_bucket.bucket
  key    = "tf-testing-obj-%[1]d-all-params"

  content             = <<CONTENT
{
  "msg": "Hi there!"
}
CONTENT
  content_type        = "application/unknown"
  cache_control       = "no-cache"
  content_disposition = "attachment"
  content_encoding    = "identity"
  content_language    = "en-GB"

  tags = {
    Key1 = "Value 1"
  }
}

data "lyvecloud_s3_object" "obj" {
  bucket = lyvecloud_s3_bucket.object_bucket.bucket
  key    = lyvecloud_s3_object.object.key
}
`, randInt)
}
