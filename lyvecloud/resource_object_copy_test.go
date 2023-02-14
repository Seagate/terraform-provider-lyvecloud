package lyvecloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccS3ObjectCopy_basic(t *testing.T) {
	rName1 := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())
	rName2 := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())
	resourceName := "lyvecloud_s3_object_copy.test"
	sourceName := "lyvecloud_s3_object.source"
	key := "HundBegraven"
	sourceKey := "WshngtnNtnls"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckObjectCopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectCopyConfig_basic(rName1, sourceKey, rName2, key),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectCopyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket", rName2),
					resource.TestCheckResourceAttr(resourceName, "key", key),
					resource.TestCheckResourceAttr(resourceName, "source", fmt.Sprintf("%s/%s", rName1, sourceKey)),
					resource.TestCheckResourceAttrPair(resourceName, "etag", sourceName, "etag"),
				),
			},
		},
	})
}

func testAccCheckObjectCopyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(Client).S3Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "lyvecloud_s3_object" {
			continue
		}

		_, err := conn.HeadObject(
			&s3.HeadObjectInput{
				Bucket:  aws.String(rs.Primary.Attributes["bucket"]),
				Key:     aws.String(rs.Primary.Attributes["key"]),
				IfMatch: aws.String(rs.Primary.Attributes["etag"]),
			})
		if err == nil {
			return fmt.Errorf("Lyve Cloud S3 Object still exists: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckObjectCopyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Object ID is set")
		}

		conn := testAccProvider.Meta().(Client).S3Client
		_, err := conn.GetObject(
			&s3.GetObjectInput{
				Bucket:  aws.String(rs.Primary.Attributes["bucket"]),
				Key:     aws.String(rs.Primary.Attributes["key"]),
				IfMatch: aws.String(rs.Primary.Attributes["etag"]),
			})
		if err != nil {
			return fmt.Errorf("S3 Object error: %s", err)
		}

		return nil
	}
}

func testAccObjectCopyConfig_basic(rName1, sourceKey, rName2, key string) string {
	return fmt.Sprintf(`
provider "lyvecloud" {
	s3 {}
}

resource "lyvecloud_s3_bucket" "source" {
  bucket = %[1]q
}

resource "lyvecloud_s3_object" "source" {
  bucket  = lyvecloud_s3_bucket.source.bucket
  key     = %[2]q
  content = "Ingen ko på isen"
}

resource "lyvecloud_s3_bucket" "target" {
  bucket = %[3]q
}

resource "lyvecloud_s3_object_copy" "test" {
  bucket = lyvecloud_s3_bucket.target.bucket
  key    = %[4]q
  source = "${lyvecloud_s3_bucket.source.bucket}/${lyvecloud_s3_object.source.key}"
}
`, rName1, sourceKey, rName2, key)
}
