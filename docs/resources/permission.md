---
page_title: "Lyve Cloud: lyvecloud_permission"
subcategory: "Account API"
description: |-
  Provides a permission resource.
---

# Resource: lyvecloud_permission

Provides a permission resource. Based on Account API.

~> **NOTE:** Credentials for Account API must be provided to use this resource.

## Example Usage

### Permission

Creating permission for multiple buckets.
```terraform
resource "lyvecloud_permission" "permission" {
  name = "my-tf-test-permission"
  description = "permission description"
  actions = "all-operations" // “all-operations”, “read-only”, or “write-only”.
  buckets = ["my-tf-test-bucket1", "my-tf-test-bucket2"]
}
```

Creating permission from a policy file.
```terraform
resource "lyvecloud_permission" "policy-permission" {
  name = "my-tf-test-policy-permission"
  description = "from policy file"
  policy = "${file("policy.json")}"
}
```

Creating a policy permission by specifying the JSON string.
```terraform
resource "lyvecloud_permission" "policy-permission" {
  name = "my-tf-test-policy-permission"
  description = "from policy file"
  policy = jsonencode({
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "statement1",
      "Action": [
        "s3:ListBucket"
      ],
      "Effect": "Allow",
      "Resource": [
        "arn:aws:s3:::mybucket"
      ],
      "Condition": {
        "StringLike": {
          "s3:prefix": [
            "David/*"
          ]
        }
      }
    },
    {
      "Sid": "statement2",
      "Action": [
        "s3:GetObject",
        "s3:PutObject"
      ],
      "Effect": "Allow",
      "Resource": [
        "arn:aws:s3:::mybucket/David/*"
      ]
    },
    {
      "Sid": "statement3",
      "Action": [
        "s3:DeleteObject"
      ],
      "Effect": "Deny",
      "Resource": [
        "arn:aws:s3:::mybucket/David/*",
        "arn:aws:s3:::mycorporatebucket/share/marketing/*"
      ]
    }
  ]
})
}
```


## Argument Reference

The following arguments are supported:

* `name` - (Optional) Specifies a unique permission name. The name allows only alphanumeric, '-', '_' or spaces Maximum length can be 128 characters. If omitted, Terraform will assign a random, unique name.
* `name_prefix` - (Optional) Creates a unique permission name beginning with the specified prefix. Conflicts with `name`.
* `description` - (Required) Description of the permission.
* `actions` - (Optional) Actions Enum: “all-operations”, “read-only”, or “write-only”. Must be set if permission is created with `buckets`, `bucket_prefix` or `all_buckets`.
Conflicts with `policy`.
* `buckets` - (Optional) List (one or more) of existing bucket names. To list one or more existing buckets you can specify 
[“bucket1”, “bucket2”, and so on]. Conflicts with `all_buckets`, `bucket_prefix` and `policy`. Required with `actions`.
* `all_buckets` - (Optional) If set to `true`, the permission is applied to all the existing and new buckets in the account. Required with `actions`. Conflicts with `buckets`, `bucket_prefix` and `policy`.
* `bucket_prefix` - (Optional) Specify the initial name of the bucket as a prefix to apply for permission. Required with `actions`. Conflicts with `buckets`, `all_buckets` and `policy`.
* `policy` - (Optional) specify a JSON file path compatible with the AWS IAM policy file or specify the JSON string as shown in the example above. Conflicts with `buckets`, `bucket_prefix`, `all_buckets` and `actions`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A Permission ID that uniquely identifies each permission created in Lyve Cloud. Can be used to identify this permission when creating a service account.
* `type` - The permission type: all-buckets/bucket-prefix/bucket-names/policy.
* `ready_state` - True if the permission is ready across all regions.