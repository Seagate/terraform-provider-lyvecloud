---
page_title: "Lyve Cloud: lyvecloud_s3_bucket"
subcategory: "S3"
description: |-
  Provides a S3 bucket resource.
---

# Resource: lyvecloud_s3_bucket

Provides a S3 bucket resource.

## Example Usage

### Bucket w/ Tags

```terraform
resource "lyvecloud_s3_bucket" "b" {
  bucket = "my-tf-test-bucket"

  tags = {
    Name        = "My bucket"
    Environment = "Dev"
  }
}
```

### Using object lock configuration

To **enable** Object Lock on a **new** bucket, use the `object_lock_enabled` argument in **this** resource.
To configure the default retention rule of the Object Lock configuration use the resource [`s3_bucket_object_lock_configuration` resource](s3_bucket_object_lock_configuration.md).

```terraform
resource "lyvecloud_s3_bucket" "example" {
  bucket = "my-tf-example-bucket"

  object_lock_enabled = true
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required, Forces new resource) The name of the bucket. Must be lowercase and less than or equal to 63 characters in length.
* `object_lock_enabled` - (Optional, Default:`false`, Forces new resource) Indicates whether this bucket has an Object Lock configuration enabled. Valid values are `true` or `false`.
* `tags` - (Optional) A map of tags to assign to the bucket.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the bucket.
* `region` - The Lyve Cloud region this bucket resides in.

## Timeouts

[Configuration options](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts):

- `create` - (Default `20m`)
- `read` - (Default `20m`)
- `update` - (Default `20m`)
- `delete` - (Default `60m`)

## Import

S3 bucket can be imported using the `bucket`, e.g.,

```
$ terraform import lyve_s3_bucket.bucket bucket-name
```