package lyvecloud

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]func() (*schema.Provider, error)
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]func() (*schema.Provider, error){
		"lyvecloud": func() (*schema.Provider, error) {
			return testAccProvider, nil
		},
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

func testAccPreCheck(t *testing.T) {
	ok := os.Getenv("TF_ACC") == "1"

	if os.Getenv("LYVECLOUD_S3_REGION") != "" {
		ok = true
	}
	if os.Getenv("LYVECLOUD_S3_ACCESS_KEY") != "" {
		ok = true
	}
	if os.Getenv("LYVECLOUD_S3_SECRET_KEY") != "" {
		ok = true
	}
	if os.Getenv("LYVECLOUD_S3_ENDPOINT") != "" {
		ok = true
	}
	if os.Getenv("LYVECLOUD_AAPIV1_CLIENT_ID") != "" {
		ok = true
	}
	if os.Getenv("LYVECLOUD_AAPIV1_CLIENT_SECRET") != "" {
		ok = true
	}
	if os.Getenv("LYVECLOUD_S3_ACCESS_KEY") != "" {
		ok = true
	}
	if os.Getenv("LYVECLOUD_S3_SECRET_KEY") != "" {
		ok = true
	}
	if os.Getenv("LYVECLOUD_S3_REGION") != "" {
		ok = true
	}
	if !ok {
		panic("you must to set env variables for integration tests!")
	}
}
