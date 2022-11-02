---
page_title: "Lyve Cloud: lyvecloud_s3_bucket_object_lock_configuration"
subcategory: "S3"
description: |-
  Provides an S3 bucket Object Lock configuration resource.
---

# Resource: lyvecloud_s3_bucket_object_lock_configuration

Provides an S3 bucket Object Lock configuration resource, known in the Lyve Cloud console as **Object Immutability**. For more information about Object Locking/Object Immutability, go to [Using object immutability](https://help.lyvecloud.seagate.com/en/using-object-immutability.html) in the Lyve Cloud Administrator's Guide.

~> **NOTE:** This resource **does not enable** Object Lock for **new** buckets. It configures a default retention period for objects placed in the specified bucket.
Thus, to **enable** Object Lock for a **new** bucket, see the [Using object lock configuration](s3_bucket.md#Using-object-lock-configuration) section in  the `lyvecloud_s3_bucket` resource or the [Object Lock configuration for a new bucket](#object-lock-configuration-for-a-new-bucket) example below.


## Example Usage

### Object Lock configuration for a new bucket

```terraform
resource "lyvecloud_s3_bucket" "example" {
  bucket = "mybucket"

  object_lock_enabled = true
}

resource "lyvecloud_s3_bucket_object_lock_configuration" "example" {
  bucket = lyvecloud_s3_bucket.example.bucket

  rule {
    default_retention {
      mode = "COMPLIANCE"
      days = 5
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required, Forces new resource) The name of the bucket.
* `object_lock_enabled` - (Optional, Forces new resource) Indicates whether this bucket has an Object Lock configuration enabled. Defaults to `Enabled`. Valid values: `Enabled`.
* `rule` - (Required) Configuration block for specifying the Object Lock rule for the specified object [detailed below](#rule).

### rule

The `rule` configuration block supports the following arguments:

* `default_retention` - (Required) A configuration block for specifying the default Object Lock retention settings for new objects placed in the specified bucket [detailed below](#default_retention).

### default_retention

The `default_retention` configuration block supports the following arguments:

* `days` - (Optional, Required if `years` is not specified) The number of days that you want to specify for the default retention period.
* `mode` - (Required) The default Object Lock retention mode you want to apply to new objects placed in the specified bucket. Valid values: `COMPLIANCE`, `GOVERNANCE`.
* `years` - (Optional, Required if `days` is not specified) The number of years that you want to specify for the default retention period.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The `bucket`.

## Import

S3 bucket Object Lock configuration can be imported using the following example command.

```
$ terraform import lyvecloud_s3_bucket_object_lock_configuration.example bucket-name
```
