package lyvecloud

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
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
)

func TestAccS3Object_noNameNoKey(t *testing.T) {
	bucketError := regexp.MustCompile(`bucket must not be empty`)
	keyError := regexp.MustCompile(`key must not be empty`)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckObjectDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig:   func() {},
				Config:      testAccObjectConfig_basic("", "a key"),
				ExpectError: bucketError,
			},
			{
				PreConfig:   func() {},
				Config:      testAccObjectConfig_basic("a name", ""),
				ExpectError: keyError,
			},
		},
	})
}

func TestAccS3Object_empty(t *testing.T) {
	var obj s3.GetObjectOutput
	resourceName := "lyvecloud_s3_object.object"
	rName := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckObjectDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccObjectConfig_empty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &obj),
					testAccCheckObjectBody(&obj, ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
				ImportStateId:           fmt.Sprintf("s3://%s/test-key", rName),
			},
		},
	})
}

func TestAccS3Object_source(t *testing.T) {
	var obj s3.GetObjectOutput
	resourceName := "lyvecloud_s3_object.object"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	source := testAccObjectCreateTempFile(t, "{anything will do }")
	defer os.Remove(source)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_source(rName, source),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &obj),
					testAccCheckObjectBody(&obj, "{anything will do }"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source", "force_destroy"},
				ImportStateId:           fmt.Sprintf("s3://%s/test-key", rName),
			},
		},
	})
}

func TestAccS3Object_content(t *testing.T) {
	var obj s3.GetObjectOutput
	resourceName := "lyvecloud_s3_object.object"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckObjectDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccObjectConfig_content(rName, "some_bucket_content"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &obj),
					testAccCheckObjectBody(&obj, "some_bucket_content"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"content", "content_base64", "force_destroy"},
				ImportStateId:           fmt.Sprintf("s3://%s/test-key", rName),
			},
		},
	})
}

func TestAccS3Object_etagEncryption(t *testing.T) {
	var obj s3.GetObjectOutput
	resourceName := "lyvecloud_s3_object.object"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	source := testAccObjectCreateTempFile(t, "{anything will do }")
	defer os.Remove(source)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckObjectDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccObjectConfig_etagEncryption(rName, source),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &obj),
					testAccCheckObjectBody(&obj, "{anything will do }"),
					resource.TestCheckResourceAttr(resourceName, "etag", "7b006ff4d70f68cc65061acf2f802e6f"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source", "force_destroy"},
				ImportStateId:           fmt.Sprintf("s3://%s/test-key", rName),
			},
		},
	})
}

func TestAccS3Object_contentBase64(t *testing.T) {
	var obj s3.GetObjectOutput
	resourceName := "lyvecloud_s3_object.object"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckObjectDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccObjectConfig_contentBase64(rName, base64.StdEncoding.EncodeToString([]byte("some_bucket_content"))),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &obj),
					testAccCheckObjectBody(&obj, "some_bucket_content"),
				),
			},
		},
	})
}

func TestAccS3Object_sourceHashTrigger(t *testing.T) {
	var obj, updated_obj s3.GetObjectOutput
	resourceName := "lyvecloud_s3_object.object"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	startingData := "Ebben!"
	changingData := "Ne andrò lontana"

	filename := testAccObjectCreateTempFile(t, startingData)
	defer os.Remove(filename)

	rewriteFile := func(*terraform.State) error {
		if err := os.WriteFile(filename, []byte(changingData), 0644); err != nil {
			os.Remove(filename)
			t.Fatal(err)
		}
		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckObjectDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccObjectConfig_sourceHashTrigger(rName, filename),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &obj),
					testAccCheckObjectBody(&obj, "Ebben!"),
					resource.TestCheckResourceAttr(resourceName, "source_hash", "7c7e02a79f28968882bb1426c8f8bfc6"),
					rewriteFile,
				),
				ExpectNonEmptyPlan: true,
			},
			{
				PreConfig: func() {},
				Config:    testAccObjectConfig_sourceHashTrigger(rName, filename),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &updated_obj),
					testAccCheckObjectBody(&updated_obj, "Ne andrò lontana"),
					resource.TestCheckResourceAttr(resourceName, "source_hash", "cffc5e20de2d21764145b1124c9b337b"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"content", "content_base64", "force_destroy", "source", "source_hash"},
				ImportStateId:           fmt.Sprintf("s3://%s/test-key", rName),
			},
		},
	})
}

func TestAccS3Object_withContentCharacteristics(t *testing.T) {
	var obj s3.GetObjectOutput
	resourceName := "lyvecloud_s3_object.object"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	source := testAccObjectCreateTempFile(t, "{anything will do }")
	defer os.Remove(source)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_contentCharacteristics(rName, source),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &obj),
					testAccCheckObjectBody(&obj, "{anything will do }"),
					resource.TestCheckResourceAttr(resourceName, "content_type", "binary/octet-stream"),
				),
			},
		},
	})
}

func TestAccS3Object_nonVersioned(t *testing.T) {
	sourceInitial := testAccObjectCreateTempFile(t, "initial object state")
	defer os.Remove(sourceInitial)
	rName := acctest.RandomWithPrefix("tf-acc-test")
	var originalObj s3.GetObjectOutput
	resourceName := "lyvecloud_s3_object.object"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_nonVersioned(rName, sourceInitial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &originalObj),
					testAccCheckObjectBody(&originalObj, "initial object state"),
					resource.TestCheckResourceAttr(resourceName, "version_id", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source", "force_destroy"},
				ImportStateId:           fmt.Sprintf("s3://%s/updateable-key", rName),
			},
		},
	})
}

func TestAccS3Object_updates(t *testing.T) {
	var originalObj, modifiedObj s3.GetObjectOutput
	resourceName := "lyvecloud_s3_object.object"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	sourceInitial := testAccObjectCreateTempFile(t, "initial object state")
	defer os.Remove(sourceInitial)
	sourceModified := testAccObjectCreateTempFile(t, "modified object")
	defer os.Remove(sourceInitial)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_updateable(rName, false, sourceInitial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &originalObj),
					testAccCheckObjectBody(&originalObj, "initial object state"),
					resource.TestCheckResourceAttr(resourceName, "etag", "647d1d58e1011c743ec67d5e8af87b53"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
			{
				Config: testAccObjectConfig_updateable(rName, false, sourceModified),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &modifiedObj),
					testAccCheckObjectBody(&modifiedObj, "modified object"),
					resource.TestCheckResourceAttr(resourceName, "etag", "1c7fd13df1515c2a13ad9eb068931f09"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source", "force_destroy"},
				ImportStateId:           fmt.Sprintf("s3://%s/updateable-key", rName),
			},
		},
	})
}

func TestAccS3Object_updateSameFile(t *testing.T) {
	var originalObj, modifiedObj s3.GetObjectOutput
	resourceName := "lyvecloud_s3_object.object"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	startingData := "lane 8"
	changingData := "chicane"

	filename := testAccObjectCreateTempFile(t, startingData)
	defer os.Remove(filename)

	rewriteFile := func(*terraform.State) error {
		if err := os.WriteFile(filename, []byte(changingData), 0644); err != nil {
			os.Remove(filename)
			t.Fatal(err)
		}
		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_updateable(rName, false, filename),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &originalObj),
					testAccCheckObjectBody(&originalObj, startingData),
					resource.TestCheckResourceAttr(resourceName, "etag", "aa48b42f36a2652cbee40c30a5df7d25"),
					rewriteFile,
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccObjectConfig_updateable(rName, false, filename),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &modifiedObj),
					testAccCheckObjectBody(&modifiedObj, changingData),
					resource.TestCheckResourceAttr(resourceName, "etag", "fafc05f8c4da0266a99154681ab86e8c"),
				),
			},
		},
	})
}

func TestAccS3Object_updatesWithVersioning(t *testing.T) {
	var originalObj, modifiedObj s3.GetObjectOutput
	resourceName := "lyvecloud_s3_object.object"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	sourceInitial := testAccObjectCreateTempFile(t, "initial versioned object state")
	defer os.Remove(sourceInitial)
	sourceModified := testAccObjectCreateTempFile(t, "modified versioned object")
	defer os.Remove(sourceInitial)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_updateable(rName, true, sourceInitial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &originalObj),
					testAccCheckObjectBody(&originalObj, "initial versioned object state"),
					resource.TestCheckResourceAttr(resourceName, "etag", "cee4407fa91906284e2a5e5e03e86b1b"),
				),
			},
			{
				Config: testAccObjectConfig_updateable(rName, true, sourceModified),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &modifiedObj),
					testAccCheckObjectBody(&modifiedObj, "modified versioned object"),
					resource.TestCheckResourceAttr(resourceName, "etag", "00b8c73b1b50e7cc932362c7225b8e29"),
					testAccCheckObjectVersionIdDiffers(&modifiedObj, &originalObj),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source", "force_destroy"},
				ImportStateId:           fmt.Sprintf("s3://%s/updateable-key", rName),
			},
		},
	})
}

func TestAccS3Object_objectLockRetentionStartWithNone(t *testing.T) {
	var obj1, obj2, obj3 s3.GetObjectOutput
	resourceName := "lyvecloud_s3_object.object"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	retainUntilDate := time.Now().UTC().AddDate(0, 0, 10).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_noLockRetention(rName, "stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &obj1),
					testAccCheckObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
			{
				Config: testAccObjectConfig_lockRetention(rName, "stuff", retainUntilDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &obj2),
					testAccCheckObjectVersionIdEquals(&obj2, &obj1),
					testAccCheckObjectBody(&obj2, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", "GOVERNANCE"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", retainUntilDate),
				),
			},
			// Remove retention period but create a new object version to test force_destroy
			{
				Config: testAccObjectConfig_noLockRetention(rName, "changed stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &obj3),
					testAccCheckObjectVersionIdDiffers(&obj3, &obj2),
					testAccCheckObjectBody(&obj3, "changed stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
		},
	})
}

func TestAccS3Object_objectLockRetentionStartWithSet(t *testing.T) {
	var obj1, obj2, obj3, obj4 s3.GetObjectOutput
	resourceName := "lyvecloud_s3_object.object"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	retainUntilDate1 := time.Now().UTC().AddDate(0, 0, 20).Format(time.RFC3339)
	retainUntilDate2 := time.Now().UTC().AddDate(0, 0, 30).Format(time.RFC3339)
	retainUntilDate3 := time.Now().UTC().AddDate(0, 0, 10).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_lockRetention(rName, "stuff", retainUntilDate1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &obj1),
					testAccCheckObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", "GOVERNANCE"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", retainUntilDate1),
				),
			},
			{
				Config: testAccObjectConfig_lockRetention(rName, "stuff", retainUntilDate2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &obj2),
					testAccCheckObjectVersionIdEquals(&obj2, &obj1),
					testAccCheckObjectBody(&obj2, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", "GOVERNANCE"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", retainUntilDate2),
				),
			},
			{
				Config: testAccObjectConfig_lockRetention(rName, "stuff", retainUntilDate3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &obj3),
					testAccCheckObjectVersionIdEquals(&obj3, &obj2),
					testAccCheckObjectBody(&obj3, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", "GOVERNANCE"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", retainUntilDate3),
				),
			},
			{
				Config: testAccObjectConfig_noLockRetention(rName, "stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &obj4),
					testAccCheckObjectVersionIdEquals(&obj4, &obj3),
					testAccCheckObjectBody(&obj4, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "object_lock_mode", ""),
					resource.TestCheckResourceAttr(resourceName, "object_lock_retain_until_date", ""),
				),
			},
		},
	})
}

func TestAccS3Object_metadata(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	var obj s3.GetObjectOutput
	resourceName := "lyvecloud_s3_object.object"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckObjectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_metadata(rName, "key1", "value1", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, "metadata.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "metadata.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "metadata.key2", "value2"),
				),
			},
			{
				Config: testAccObjectConfig_metadata(rName, "key1", "value1updated", "key3", "value3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, "metadata.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "metadata.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "metadata.key3", "value3"),
				),
			},
			{
				Config: testAccObjectConfig_empty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &obj),
					resource.TestCheckResourceAttr(resourceName, "metadata.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
				ImportStateId:           fmt.Sprintf("s3://%s/test-key", rName),
			},
		},
	})
}

func TestAccS3Object_tags(t *testing.T) {
	var obj1, obj2, obj3, obj4 s3.GetObjectOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "lyvecloud_s3_object.object"
	key := "test-key"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckObjectDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config:    testAccObjectConfig_tags(rName, key, "stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &obj1),
					testAccCheckObjectBody(&obj1, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "A@AA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
			{
				PreConfig: func() {},
				Config:    testAccObjectConfig_updatedTags(rName, key, "stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &obj2),
					testAccCheckObjectVersionIdEquals(&obj2, &obj1),
					testAccCheckObjectBody(&obj2, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "B@BB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "X X"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key4", "DDD"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key5", "E:/"),
				),
			},
			{
				PreConfig: func() {},
				Config:    testAccObjectConfig_noTags(rName, key, "stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &obj3),
					testAccCheckObjectVersionIdEquals(&obj3, &obj2),
					testAccCheckObjectBody(&obj3, "stuff"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				PreConfig: func() {},
				Config:    testAccObjectConfig_tags(rName, key, "changed stuff"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectExists(resourceName, &obj4),
					testAccCheckObjectVersionIdDiffers(&obj4, &obj3),
					testAccCheckObjectBody(&obj4, "changed stuff"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "A@AA"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "BBB"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "CCC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"content", "force_destroy"},
				ImportStateId:           fmt.Sprintf("s3://%s/%s", rName, key),
			},
		},
	})
}

func testAccCheckObjectDestroy(s *terraform.State) error {
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

func testAccCheckObjectExists(n string, obj *s3.GetObjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Object ID is set")
		}

		conn := testAccProvider.Meta().(Client).S3Client

		input := &s3.GetObjectInput{
			Bucket:  aws.String(rs.Primary.Attributes["bucket"]),
			Key:     aws.String(rs.Primary.Attributes["key"]),
			IfMatch: aws.String(rs.Primary.Attributes["etag"]),
		}

		var out *s3.GetObjectOutput

		err := resource.RetryContext(context.Background(), 2*time.Minute, func() *resource.RetryError {
			var err error
			out, err = conn.GetObject(input)

			if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchKey) {
				return resource.RetryableError(
					fmt.Errorf("getting object %s, retrying: %w", rs.Primary.Attributes["bucket"], err),
				)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})
		if TimedOut(err) {
			out, err = conn.GetObject(input)
		}

		if err != nil {
			return fmt.Errorf("S3 Object error: %s", err)
		}

		*obj = *out

		return nil
	}
}

func testAccCheckObjectBody(obj *s3.GetObjectOutput, want string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		body, err := io.ReadAll(obj.Body)
		if err != nil {
			return fmt.Errorf("failed to read body: %s", err)
		}
		obj.Body.Close()

		if got := string(body); got != want {
			return fmt.Errorf("wrong result body %q; want %q", got, want)
		}

		return nil
	}
}

func testAccObjectCreateTempFile(t *testing.T, data string) string {
	tmpFile, err := os.CreateTemp("", "tf-acc-s3-obj")
	if err != nil {
		t.Fatal(err)
	}
	filename := tmpFile.Name()

	err = os.WriteFile(filename, []byte(data), 0644)
	if err != nil {
		os.Remove(filename)
		t.Fatal(err)
	}

	return filename
}

func testAccCheckObjectVersionIdEquals(first, second *s3.GetObjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if first.VersionId == nil {
			return fmt.Errorf("Expected first object to have VersionId: %s", first)
		}
		if second.VersionId == nil {
			return fmt.Errorf("Expected second object to have VersionId: %s", second)
		}

		if *first.VersionId != *second.VersionId {
			return fmt.Errorf("Expected Version IDs to be equal, but they differ (%s, %s)", *first.VersionId, *second.VersionId)
		}

		return nil
	}
}

func testAccCheckObjectVersionIdDiffers(first, second *s3.GetObjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if first.VersionId == nil {
			return fmt.Errorf("Expected first object to have VersionId: %s", first)
		}
		if second.VersionId == nil {
			return fmt.Errorf("Expected second object to have VersionId: %s", second)
		}

		if *first.VersionId == *second.VersionId {
			return fmt.Errorf("Expected Version IDs to differ, but they are equal (%s)", *first.VersionId)
		}

		return nil
	}
}

func testAccObjectConfig_basic(bucket, key string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_object" "object" {
  bucket = %[1]q
  key    = %[2]q
}
`, bucket, key)
}

func testAccObjectConfig_empty(rName string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "test" {
  bucket = %[1]q
}

resource "lyvecloud_s3_object" "object" {
  bucket = lyvecloud_s3_bucket.test.bucket
  key    = "test-key"
}
`, rName)
}

func testAccObjectConfig_source(rName string, source string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "test" {
  bucket = %[1]q
}

resource "lyvecloud_s3_object" "object" {
  bucket       = lyvecloud_s3_bucket.test.bucket
  key          = "test-key"
  source       = %[2]q
  content_type = "binary/octet-stream"
}
`, rName, source)
}

func testAccObjectConfig_content(rName string, content string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "test" {
  bucket = %[1]q
}

resource "lyvecloud_s3_object" "object" {
  bucket  = lyvecloud_s3_bucket.test.bucket
  key     = "test-key"
  content = %[2]q
}
`, rName, content)
}

func testAccObjectConfig_etagEncryption(rName string, source string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "test" {
  bucket = %[1]q
}

resource "lyvecloud_s3_object" "object" {
  bucket                 = lyvecloud_s3_bucket.test.bucket
  key                    = "test-key"
  source                 = %[2]q
  etag                   = filemd5(%[2]q)
}
`, rName, source)
}

func testAccObjectConfig_contentBase64(rName string, contentBase64 string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "test" {
  bucket = %[1]q
}

resource "lyvecloud_s3_object" "object" {
  bucket         = lyvecloud_s3_bucket.test.bucket
  key            = "test-key"
  content_base64 = %[2]q
}
`, rName, contentBase64)
}

func testAccObjectConfig_sourceHashTrigger(rName string, source string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "test" {
  bucket = %[1]q
}

resource "lyvecloud_s3_object" "object" {
  bucket      = lyvecloud_s3_bucket.test.bucket
  key         = "test-key"
  source      = %[2]q
  source_hash = filemd5(%[2]q)
}
`, rName, source)
}

func testAccObjectConfig_contentCharacteristics(rName string, source string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "test" {
  bucket = %[1]q
}

resource "lyvecloud_s3_object" "object" {
  bucket           = lyvecloud_s3_bucket.test.bucket
  key              = "test-key"
  source           = %[2]q
  content_language = "en"
  content_type     = "binary/octet-stream"
}
`, rName, source)
}

func testAccObjectConfig_nonVersioned(rName string, source string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "object_bucket_3" {
  bucket = %[1]q
}

resource "lyvecloud_s3_object" "object" {
  bucket = lyvecloud_s3_bucket.object_bucket_3.bucket
  key    = "updateable-key"
  source = %[2]q
  etag   = filemd5(%[2]q)
}
`, rName, source)
}

func testAccObjectConfig_updateable(rName string, bucketVersioning bool, source string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "object_bucket_3" {
  bucket = %[1]q
  object_lock_enabled = true
}

resource "lyvecloud_s3_object" "object" {
  # Must have bucket versioning enabled first
  bucket = lyvecloud_s3_bucket.object_bucket_3.bucket
  key    = "updateable-key"
  source = %[3]q
  etag   = filemd5(%[3]q)
}
`, rName, bucketVersioning, source)
}

func testAccObjectConfig_metadata(rName string, metadataKey1, metadataValue1, metadataKey2, metadataValue2 string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "test" {
  bucket = %[1]q
}

resource "lyvecloud_s3_object" "object" {
  bucket = lyvecloud_s3_bucket.test.bucket
  key    = "test-key"

  metadata = {
    %[2]s = %[3]q
    %[4]s = %[5]q
  }
}
`, rName, metadataKey1, metadataValue1, metadataKey2, metadataValue2)
}

func testAccObjectConfig_tags(rName, key, content string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "test" {
  bucket = %[1]q
  object_lock_enabled = true
}

resource "lyvecloud_s3_object" "object" {
  bucket  = lyvecloud_s3_bucket.test.bucket
  key     = %[2]q
  content = %[3]q

  tags = {
    Key1 = "A@AA"
    Key2 = "BBB"
    Key3 = "CCC"
  }
}
`, rName, key, content)
}

func testAccObjectConfig_updatedTags(rName, key, content string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "test" {
  bucket = %[1]q
  object_lock_enabled = true
}

resource "lyvecloud_s3_object" "object" {
  bucket  = lyvecloud_s3_bucket.test.bucket
  key     = %[2]q
  content = %[3]q

  tags = {
    Key2 = "B@BB"
    Key3 = "X X"
    Key4 = "DDD"
    Key5 = "E:/"
  }
}
`, rName, key, content)
}

func testAccObjectConfig_noTags(rName, key, content string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "test" {
  bucket = %[1]q
  object_lock_enabled = true
}

resource "lyvecloud_s3_object" "object" {
  bucket  = lyvecloud_s3_bucket.test.bucket
  key     = %[2]q
  content = %[3]q
}
`, rName, key, content)
}

func testAccObjectConfig_noLockRetention(rName string, content string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "test" {
  bucket = %[1]q

  object_lock_enabled = true
}

resource "lyvecloud_s3_object" "object" {
  bucket        = lyvecloud_s3_bucket.test.bucket
  key           = "test-key"
  content       = %[2]q
  force_destroy = true
}
`, rName, content)
}

func testAccObjectConfig_lockRetention(rName string, content, retainUntilDate string) string {
	return fmt.Sprintf(`
resource "lyvecloud_s3_bucket" "test" {
  bucket = %[1]q

  object_lock_enabled = true
}

resource "lyvecloud_s3_object" "object" {
  bucket                        = lyvecloud_s3_bucket.test.bucket
  key                           = "test-key"
  content                       = %[2]q
  force_destroy                 = true
  object_lock_mode              = "GOVERNANCE"
  object_lock_retain_until_date = %[3]q
}
`, rName, content, retainUntilDate)
}
