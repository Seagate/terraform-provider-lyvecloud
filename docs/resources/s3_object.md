---
page_title: "Lyve Cloud: lyvecloud_s3_object"
subcategory: "S3"
description: |-
  Provides a S3 object resource.
---

# Resource: lyvecloud_s3_object

Provides an S3 object resource.

## Example Usage

### Uploading a file to a bucket

```terraform
resource "lyvecloud_s3_object" "object" {
  bucket = "your_bucket_name"
  key    = "new_object_key"
  source = "path/to/file"
}
```

## Argument Reference

-> **Note:** If you specify `content_encoding` you are responsible for encoding the body appropriately. `source`, `content`, and `content_base64` all expect already encoded/compressed bytes.

The following arguments are required:

* `bucket` - (Required) Name of the bucket to put the file in.
* `key` - (Required) Name of the object once it is in the bucket.


The following arguments are optional:

* `source` - (Required) Path to a file that will be read and uploaded as raw bytes for the object content.
* `cache_control` - (Optional) Caching behavior along the request/reply chain Read [w3c cache_control](http://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.9) for further details.
* `content_base64` - (Optional, conflicts with `source` and `content`) Base64-encoded data that will be decoded and uploaded as raw bytes for the object content. This allows safely uploading non-UTF8 binary data, but is recommended only for small content such as the result of the `gzipbase64` function with small text strings. For larger objects, use `source` to stream the content from a disk file.
* `content_disposition` - (Optional) Presentational information for the object. Read [w3c content_disposition](http://www.w3.org/Protocols/rfc2616/rfc2616-sec19.html#sec19.5.1) for further information.
* `content_encoding` - (Optional) Content encodings that have been applied to the object and thus what decoding mechanisms must be applied to obtain the media-type referenced by the Content-Type header field.
* `content_language` - (Optional) Language the content is in e.g., en-US or en-GB.
* `content_type` - (Optional) Standard MIME type describing the format of the object data, e.g., application/octet-stream. All Valid MIME Types are valid for this input.
* `content` - (Optional, conflicts with `source` and `content_base64`) Literal string value to use as the object content, which will be uploaded as UTF-8-encoded text.
* `etag` - (Optional) Triggers updates when the value changes. The only meaningful value is `filemd5("path/to/file")` (Terraform 0.11.12 or later) or `${md5(file("path/to/file"))}` (Terraform 0.11.11 or earlier).
* `force_destroy` - (Optional) Whether to allow the object to be deleted by removing any legal hold on any object version. Default is `false`. This value should be set to `true` only if the bucket has S3 object lock enabled. **Note** that objects with retention mode *COMPLIANCE* will not be affected by this flag.
* `metadata` - (Optional) Map of keys/values to provision metadata (will be automatically prefixed by `x-amz-meta-`, note that only lowercase label are currently supported by the AWS Go API).
* `object_lock_mode` - (Optional) Object lock retention mode that you want to apply to this object. Valid values are `GOVERNANCE` and `COMPLIANCE`.
* `object_lock_retain_until_date` - (Optional) Date and time, in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8), when this object's object lock will expire.
* `tags` - (Optional) Map of tags to assign to the object.

If no content is provided through `source`, `content` or `content_base64`, then the object will be empty.

**Note:** Terraform ignores all leading `/`s in the object's `key` and treats multiple `/`s in the rest of the object's `key` as a single `/`, so values of `/index.html` and `index.html` correspond to the same S3 object as do `first//second///third//` and `first/second/third/`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `etag` - ETag generated for the object.
* `id` - `key` of the resource supplied above.
* `tags` - Map of tags assigned to the resource.
* `version_id` - Unique version ID value for the object, if bucket versioning is enabled.


## Import

Objects can be imported using the `id`. The `id` is the bucket name and the key together e.g.,

```
$ terraform import lyvecloud_s3_object.object some-bucket-name/some/key.txt
```

Additionally, s3 url syntax can be used, e.g.,

```
$ terraform import lyve_s3_object.object s3://some-bucket-name/some/key.txt
```
