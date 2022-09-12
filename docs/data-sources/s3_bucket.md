---
subcategory: "S3"
page_title: "Lyve Cloud: lyvecloud_s3_bucket"
description: |-
    Provides details about a specific S3 bucket
---

# lyvecloud_s3_bucket (Data Source)
Provides details about a specific S3 bucket.

## Example Usage

### Printing bucket's region

```terraform
data "aws_s3_bucket" "selected" {
  bucket = "bucket.test.com"
}

output "bucket_region" {
  value = aws_s3_bucket.selected.region
}

```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required) The name of the bucket.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the bucket.
* `region` - The Lyve Cloud region this bucket resides in.