# create service account and output its credentials
resource "lyvecloud_service_account" "test" {
	name = "tf-sa-test-1"
	description = "description is optional"
	permissions = ["permission-1-id", "permission-2-id"]
}

output "access-key" {
  value = lyvecloud_service_account.test.access_key
}

output "secret-key" {
  value = lyvecloud_service_account.test.secret
}