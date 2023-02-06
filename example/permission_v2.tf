# permission for list of buckets
resource "lyvecloud_permission_v2" "p1" {
	name = "tf-perm-test-1"
	description = "description is required"
	actions = "all-operations"
	buckets = ["my-bucket-1", "my-bucket-2"]
}

# permission for buckets prefix
resource "lyvecloud_permission_v2" "p2" {
	name = "tf-perm-test-2"
	description = "description is required"
	actions = "read-only"
	bucket_prefix = "abc-"
}

# permission for all buckets
resource "lyvecloud_permission_v2" "p3" {
	name = "tf-perm-test-3"
	description = "description is required"
	actions = "write-only"
	all_buckets = true
}

# output id of permission. used later to create service account
output "permission-id" {
  value = lyvecloud_permission_v2.test.access_key
}