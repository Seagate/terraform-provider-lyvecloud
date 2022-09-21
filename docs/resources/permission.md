---
page_title: "Lyve Cloud: lyvecloud_permission"
subcategory: "Account"
description: |-
  Provides a permission resource.
---

# Resource: lyvecloud_permission

Provides a permission resource.

## Example Usage

### Permission

```terraform
resource "lyvecloud_permission" "permission" {
  permission = "my-tf-test-permission"
  description = "permission description"
  actions = "all-operations" // “all-operations”, “read”, or “write”.
  buckets = ["my-tf-test-bucket1", "my-tf-test-bucket2"]
}
```

## Argument Reference

The following arguments are supported:

* `permission` - Specifies a unique Permission name. The name allows only alphanumeric, '-', '_' or spaces Maximum length can be 128 characters.
* `description` - Description of the permission.
* `actions` - Actions Enum: “all-operations”, “read”, or “write”.
  Terraform wil only perform drift detection if a configuration value is provided.
* `buckets` - List (one or more) of existing bucket names or add a prefix followed by asterix to specify all the buckets in the account that start with the prefix. To list one or more existing buckets you can specify 
[“bucket1”, “bucket2”, and so on]. Adding a prefix ["abc-*"] will apply permission to all the buckets with prefix "abc-".

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A Permission ID that uniquely identifies each permission created in Lyve Cloud. Can be used to identify this permission when creating a service account.