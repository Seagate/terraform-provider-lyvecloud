---
page_title: "Lyve Cloud: lyvecloud_service_account"
subcategory: "Account"
description: |-
  Provides a service account resource.
---

# Resource: lyvecloud_service_account

Provides a service account resource. Based on Account API v1.

~> **NOTE:** Updating service accounts is not supported in this resource. For the functionality of updating service accounts, 
`resource_service_account_v2` must be used.

## Example Usage

### Service Account

```terraform
resource "lyvecloud_service_account" "serviceaccount" {
  service_account = "my-tf-test-service_account"
  description = "service account description"
  permissions = ["my-tf-test-permission-id"]
}
```

## Argument Reference
The following arguments are supported:

* `service_account` - (Required) Specifies the unique Service Account name. The name allows only alphanumeric, '-', '_' or space.
* `description` - (Optional) Description of the Service Account.
* `permissions` - (Required) Specify (one or more) unique values of permission-id.

## Attributes Reference
In addition to all arguments above, the following attributes are exported:

* `id` - A Service Account ID that uniquely identifies each Service Account created in Lyve Cloud. Used to identify this Service Account when it is deleted.
* `access_key` - Access key to use when authenticating S3 API requests.
* `access_secret` - Access secret key to use when authenticating S3 API requests.