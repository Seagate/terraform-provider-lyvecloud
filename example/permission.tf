# permission for list of buckets
resource "lyvecloud_permission" "p1" {
	name = "tf-perm-test-1"
	description = "description is required"
	actions = "all-operations"
	buckets = ["my-bucket-1", "my-bucket-2"]
}

# permission for buckets prefix
resource "lyvecloud_permission" "p2" {
	name = "tf-perm-test-2"
	description = "description is required"
	actions = "read-only"
	bucket_prefix = "abc-"
}

# permission for all buckets
resource "lyvecloud_permission" "p3" {
	name = "tf-perm-test-3"
	description = "description is required"
	actions = "write-only"
	all_buckets = true
}

# permission with policy
resource "lyvecloud_permission" "p4" {
	name = "tf-perm-test-4"
	description = "from policy string"
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
      }
    ]
  })
}

# permission from policy json file
resource "lyvecloud_permission" "p5" {
	name = "tf-perm-test-5"
	description = "from policy file"
  policy = "${file("policy.json")}"
}

# output id of permission. used later to create service account
output "permission-id" {
  value = lyvecloud_permission.p1.access_key
}