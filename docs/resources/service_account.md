---
page_title: "Lyve Cloud: lyvecloud_service_account"
subcategory: "Account"
description: |-
  Provides a service account resource.
---

# Resource: lyvecloud_service_account

Provides a service account resource. Based on Account API.

~> **NOTE:** Credentials for Account API must be provided to use this resource.

## Example Usage

### Service Account

```terraform
resource "lyvecloud_service_account" "serviceaccount" {
  name = "my-tf-test-service_account"
  description = "service account description"
  permissions = ["my-tf-test-permission-id"]
}
```

## Argument Reference
The following arguments are supported:

* `name` - (Required) Specifies the unique Service Account name. The name allows only alphanumeric, '-', '_' or space.
* `description` - (Optional) Description of the Service Account.
* `permissions` - (Required) Specify (one or more) unique values of permission-id.

## Attributes Reference
In addition to all arguments above, the following attributes are exported:

* `id` - A Service Account ID that uniquely identifies each Service Account created in Lyve Cloud. Used to identify this Service Account when it is deleted.
* `access_key` - Access key to use when authenticating S3 API requests.
* `secret` - Access secret key to use when authenticating S3 API requests.
* `ready_state` - True if the service account is ready across all regions.
* `enabled` - State of the Service Account. It can be enabled or disabled.

## Import

Service Account can be imported using the `service account`, e.g.,

```
$ terraform import lyvecloud_servcie_account.servcie-account servcie-account-id
```