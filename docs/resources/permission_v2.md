---
page_title: "Lyve Cloud: lyvecloud_permission_v2"
subcategory: "Account API v2"
description: |-
  Provides a permission resource.
---

# Resource: lyvecloud_permission

Provides a permission resource. Based on Account API v2.

~> **NOTE:** Credentials for Account API v2 must be provided to use this resource.

## Example Usage

### Permission

```terraform
resource "lyvecloud_permission_v2" "permission" {
  permission = "my-tf-test-permission"
  description = "permission description"
  actions = "all-operations" // “all-operations”, “read-only”, or “write-only”.
  buckets = ["my-tf-test-bucket1", "my-tf-test-bucket2"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Specifies a unique Permission name. The name allows only alphanumeric, '-', '_' or spaces Maximum length can be 128 characters. If omitted, Terraform will assign a random, unique name.
* `name_prefix` - (Optional) Creates a unique permission name beginning with the specified prefix. Conflicts with `name`.
* `description` - (Required) Description of the permission.
* `actions` - (Required) Actions Enum: “all-operations”, “read-only”, or “write-only”.
  Terraform wil only perform drift detection if a configuration value is provided.
* `buckets` - (Optional) List (one or more) of existing bucket names. To list one or more existing buckets you can specify 
[“bucket1”, “bucket2”, and so on]. Conflicts with `all_buckets` and `bucket_prefix`.
* `all_buckets` - (Optional) If set to `true`, the permission is applied to all the existing and new buckets in the account.
* `bucket_prefix` - (Optional) Specify the initial name of the bucket as a prefix to apply for permission.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A Permission ID that uniquely identifies each permission created in Lyve Cloud. Can be used to identify this permission when creating a service account.
* `type` - The permission type: all-buckets/bucket-prefix/bucket-names/policy.
* `ready_state` - True if the permission is ready across all regions.