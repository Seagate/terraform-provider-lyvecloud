package lyvecloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
)

func TestAccCreatePermission_Basic(t *testing.T) {
	bucketName := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())
	permissionName := fmt.Sprintf("tf-test-permission-%d", acctest.RandInt())

	resourceName := "lyvecloud_permission.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testCreatePermission_Basic(bucketName, permissionName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}

func TestAccCreatePermission_Prefix(t *testing.T) {
	bucketName1 := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())
	bucketName2 := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())
	permissionName := fmt.Sprintf("tf-test-permission-%d", acctest.RandInt())

	resourceName := "lyvecloud_permission.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testCreatePermission_Prefix(bucketName1, bucketName2, permissionName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
		},
	})
}

func testAccCheckPermissionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(Client).AccAPIV1Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "lyvecloud_permission" {
			continue
		}

		_, err := conn.DeletePermission(rs.Primary.Attributes["id"])

		if err == nil {
			return fmt.Errorf("Permission still exists: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCreatePermission_Basic(bucketName, permissionName string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "test" {
	bucket = %[1]q
}

resource "lyvecloud_permission" "test" {
	name = %[2]q
	description = ""
	actions = "all-operations"
	buckets = [lyvecloud_s3_bucket.test.bucket]
}
`, bucketName, permissionName)
}

func testCreatePermission_Prefix(bucketName1, bucketName2, permissionName string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "test1" {
  bucket = %[1]q
}

resource "lyvecloud_s3_bucket" "test2" {
	bucket = %[2]q
}

resource "lyvecloud_permission" "test" {
	name = %[3]q
	description = ""
	actions = "all-operations" // “all-operations”, “read”, or “write”.
	buckets = ["tf-test-bucket*"]
  }
`, bucketName1, bucketName2, permissionName)
}
