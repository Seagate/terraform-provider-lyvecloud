---
page_title: "Lyve Cloud: lyvecloud_s3_object"
subcategory: "S3"
description: |-
    Provides metadata and optionally content of an S3 object
---

# lyvecloud_s3_object (Data Source)

The S3 object data source allows access to the metadata and
_optionally_ (see below) content of an object stored inside S3 bucket.

~> **Note:** The content of an object (`body` field) is available only for objects which have a human-readable `Content-Type` (`text/*` and `application/json`). This is to prevent printing unsafe characters and potentially downloading large amount of data which would be thrown away in favour of metadata.

## Example Usage

The following example retrieves a text object (which must have a `Content-Type`
value starting with `text/`) and prints it content:

```terraform
data "lyvecloud_s3_object" "selected" {
  bucket = "my-tf-bucket"
  key    = "text-file.txt"
}

output "print_content" {
  value = lyvecloud_s3_object.selected.body
}

```


## Argument Reference

The following arguments are supported:

* `bucket` - (Required) The name of the bucket to read the object from.
* `key` - (Required) The full path to the object inside the bucket.
* `version_id` - (Optional) Specific version ID of the object returned (defaults to latest version).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `body` - Object data (see **limitations above** to understand cases in which this field is actually available).
* `cache_control` - Specifies caching behavior along the request/reply chain.
* `content_disposition` - Specifies presentational information for the object.
* `content_encoding` - Specifies what content encodings have been applied to the object and thus what decoding mechanisms must be applied to obtain the media-type referenced by the Content-Type header field.
* `content_language` - The language the content is in.
* `content_length` - Size of the body in bytes.
* `content_type` - A standard MIME type describing the format of the object data.
* `etag` - [ETag](https://en.wikipedia.org/wiki/HTTP_ETag) generated for the object (an MD5 sum of the object content in case it's not encrypted).
* `last_modified` - Last modified date of the object in RFC1123 format (e.g., `Mon, 02 Jan 2006 15:04:05 MST`).
* `metadata` - A map of metadata stored with the object in S3.
* `object_lock_mode` - The object lock retention mode currently in place for this object.
* `object_lock_retain_until_date` - The date and time when this object's object lock will expire.
* `version_id` - The latest version ID of the object returned.
* `tags`  - A map of tags assigned to the object.

-> **Note:** Terraform ignores all leading `/`s in the object's `key` and treats multiple `/`s in the rest of the object's `key` as a single `/`, so values of `/index.html` and `index.html` correspond to the same S3 object as do `first//second///third//` and `first/second/third/`.