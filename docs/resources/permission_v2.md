---
page_title: "Lyve Cloud: lyvecloud_permission_v2"
subcategory: "Account API v2"
description: |-
  Provides a permission resource.
---

# Resource: lyvecloud_permission

Provides a permission resource. Based on Account API v2.

~> **NOTE:** Credentials for Account API v2 must be provided to use this resource.

~> **NOTE:** Policy permissions can only be updated if they remain policy permissions. If a policy permission is changed to a different type of permission, a new resource must be forced and vice versa, which may fail if the permission is associated with a service account.


## Example Usage

### Permission

Creating permission for multiple buckets.
```terraform
resource "lyvecloud_permission_v2" "permission" {
  name = "my-tf-test-permission"
  description = "permission description"
  actions = "all-operations" // “all-operations”, “read-only”, or “write-only”.
  buckets = ["my-tf-test-bucket1", "my-tf-test-bucket2"]
}
```

Creating permission from a policy file.
```terraform
resource "lyvecloud_permission_v2" "policy-permission" {
  name = "my-tf-test-policy-permission"
  description = "from policy file"
  policy = "${file("policy.json")}"
}
```

Creating a policy permission by specifying the JSON string.
```terraform
resource "lyvecloud_permission_v2" "policy-permission" {
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

* `name` - (Optional) Specifies a unique Permission name. The name allows only alphanumeric, '-', '_' or spaces Maximum length can be 128 characters. If omitted, Terraform will assign a random, unique name.
* `name_prefix` - (Optional) Creates a unique permission name beginning with the specified prefix. Conflicts with `name`.
* `description` - (Required) Description of the permission.
* `actions` - (Optional) Actions Enum: “all-operations”, “read-only”, or “write-only”.
  Terraform wil only perform drift detection if a configuration value is provided.
* `buckets` - (Optional) List (one or more) of existing bucket names. To list one or more existing buckets you can specify 
[“bucket1”, “bucket2”, and so on]. Conflicts with `all_buckets` and `bucket_prefix`. [“bucket1”, “bucket2”, and so on]. Conflicts with `all_buckets` and `bucket_prefix`. Required with `actions`.
* `all_buckets` - (Optional) If set to `true`, the permission is applied to all the existing and new buckets in the account. Required with `actions`.
* `bucket_prefix` - (Optional) Specify the initial name of the bucket as a prefix to apply for permission. Required with `actions`.
* `policy` - (Optional) specify a JSON file path compatible with the AWS IAM policy file or specify the JSON string as shown in the example above.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A Permission ID that uniquely identifies each permission created in Lyve Cloud. Can be used to identify this permission when creating a service account.
* `type` - The permission type: all-buckets/bucket-prefix/bucket-names/policy.
* `ready_state` - True if the permission is ready across all regions.