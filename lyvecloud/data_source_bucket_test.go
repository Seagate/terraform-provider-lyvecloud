package lyvecloud

import (
	"fmt"
	"os"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccS3BucketDataSource_basic(t *testing.T) {
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")
	region := os.Getenv("LYVECLOUD_S3_REGION")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccBucketDataSourceConfig_basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBucketExists("data.lyvecloud_s3_bucket.bucket"),
					resource.TestCheckResourceAttr("data.lyvecloud_s3_bucket.bucket", "region", region),
				),
			},
		},
	})
}

func testAccBucketDataSourceConfig_basic(bucketName string) string {
	return fmt.Sprintf(`
provider "lyvecloud" {
	s3 {}
}

resource "lyvecloud_s3_bucket" "bucket" {
  bucket = %[1]q
}

data "lyvecloud_s3_bucket" "bucket" {
  bucket = lyvecloud_s3_bucket.bucket.id
}
`, bucketName)
}
