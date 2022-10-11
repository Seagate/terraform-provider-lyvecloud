---
page_title: "Lyve Cloud: lyvecloud_s3_object_copy"
subcategory: "S3"
description: |-
  Provides a resource for copying an S3 object.
---

# Resource: lyvecloud_s3_object_copy

Provides a resource for copying an S3 object.

## Example Usage

```terraform
resource "lyvecloud_s3_object_copy" "test" {
  bucket = "destination_bucket"
  key    = "destination_key"
  source = "source_bucket/source_key"
}
```

## Argument Reference

The following arguments are required:

* `bucket` - (Required) Name of the bucket to put the file in.
* `key` - (Required) Name of the object once it is in the bucket.
* `source` - (Required) Specifies the source object for the copy operation. You specify the value in one of two formats. For objects not accessed through an access point, specify the name of the source bucket and the key of the source object, separated by a slash (`/`). For example, `testbucket/test1.json`.

The following arguments are optional:

* `cache_control` - (Optional) Specifies caching behavior along the request/reply chain Read [w3c cache_control](http://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.9) for further details.
* `content_disposition` - (Optional) Specifies presentational information for the object. Read [w3c content_disposition](http://www.w3.org/Protocols/rfc2616/rfc2616-sec19.html#sec19.5.1) for further information.
* `content_encoding` - (Optional) Specifies what content encodings have been applied to the object and thus what decoding mechanisms must be applied to obtain the media-type referenced by the Content-Type header field. Read [w3c content encoding](http://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.11) for further information.
* `content_language` - (Optional) Language the content is in e.g., en-US or en-GB.
* `content_type` - (Optional) Standard MIME type describing the format of the object data, e.g., `application/octet-stream`. All Valid MIME Types are valid for this input.
* `copy_if_match` - (Optional) Copies the object if its entity tag (ETag) matches the specified tag.
* `copy_if_modified_since` - (Optional) Copies the object if it has been modified since the specified time, in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `copy_if_none_match` - (Optional) Copies the object if its entity tag (ETag) is different than the specified ETag.
* `copy_if_unmodified_since` - (Optional) Copies the object if it hasn't been modified since the specified time, in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `force_destroy` - (Optional) Allow the object to be deleted by removing any legal hold on any object version. Default is `false`. This value should be set to `true` only if the bucket has S3 object lock enabled.
* `metadata` - (Optional) A map of keys/values to provision metadata (will be automatically prefixed by `x-amz-meta-`, note that only lowercase label are currently supported by the AWS Go API).
* `metadata_directive` - (Optional) Specifies whether the metadata is copied from the source object or replaced with metadata provided in the request. Valid values are `COPY` and `REPLACE`.
* `tagging_directive` - (Optional) Specifies whether the object tag-set are copied from the source object or replaced with tag-set provided in the request. Valid values are `COPY` and `REPLACE`.
* `tags` - (Optional) A map of tags to assign to the object. **Recommended** to use in combination with the `tagging_directive` argument to avoid inconsistent results.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `etag` - The ETag generated for the object (an MD5 sum of the object content).
* `id` - The `key` of the resource supplied above.
* `last_modified` - Returns the date that the object was last modified, in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8).
* `source_version_id` - Version of the copied object in the source bucket.
* `version_id` - Version ID of the newly created copy.
