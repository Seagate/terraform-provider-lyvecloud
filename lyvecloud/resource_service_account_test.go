package lyvecloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
)

func TestAccCreateServiceAccount_Basic(t *testing.T) {
	bucketName := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())
	permissionName := fmt.Sprintf("tf-test-permission-%d", acctest.RandInt())
	serviceAccountName := fmt.Sprintf("tf-test-service-account-%d", acctest.RandInt())

	resourceName := "lyvecloud_service_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckServiceAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testCreateServiceAccount_Basic(bucketName, permissionName, serviceAccountName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "access_key"),
					resource.TestCheckResourceAttrSet(resourceName, "access_secret"),
				),
			},
		},
	})
}

func testAccCheckServiceAccountDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(Client).AccAPIV1Client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "lyvecloud_service_account" {
			continue
		}

		_, err := conn.DeleteServiceAccount(rs.Primary.Attributes["id"])

		if err == nil {
			return fmt.Errorf("Service Account still exists: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testCreateServiceAccount_Basic(bucketName, permissionName, serviceAccountName string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "test" {
	bucket = %[1]q
}

resource "lyvecloud_permission" "test" {
	permission = %[2]q
	description = ""
	actions = "all-operations"
	buckets = [lyvecloud_s3_bucket.test.bucket]
}

resource "lyvecloud_service_account" "test" {
	service_account = %[3]q
	description = ""
	permissions = [lyvecloud_permission.test.id]
}
`, bucketName, permissionName, serviceAccountName)
}
