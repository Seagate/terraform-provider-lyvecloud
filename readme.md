## Terraform Provider for Lyve Cloud
<p align="center">
  <a href="https://github.com/Seagate/terraform-provider-lyvecloud">
    <img src="images/tf-lc.png" alt="lyvecloud-provider-terraform" width="200">
  </a>
  <p align="center">

A [Terraform](https://www.terraform.io) provider to manage [Lyve Cloud Storage](https://www.seagate.com/gb/en/services/cloud/storage/). \
This project is based on the official [AWS provider](https://github.com/hashicorp/terraform-provider-aws).

## Requirements

* [`Go 1.19`](https://go.dev/doc/install) (to build the provider plugin)
* [`Terraform v1.2`](https://www.terraform.io/downloads)

## Usage

### Configure
To quickly get started using the provider, configure the provider as shown below.

You can set the credentials of the S3 API to manage buckets and objects, the credentials of the Account API to manage permissions and service accounts, or both.

```hcl
provider "lyvecloud" {
  //s3 api - optional
  region = ""
  access_key = ""
  secret_key = ""
  endpoint_url = ""

  //account api - optional
  client_id = ""
  client_secret = ""
}
```


### Create bucket example

```hcl
resource "lyvecloud_s3_bucket" "bucket" {
  bucket = "my-tf-bucket"
}
```

### Create permission example

```hcl
resource "lyvecloud_permission" "permission" {
  permission = "my-tf-permission"
  description = "my-example-permission-description"
  actions = "all-operations" // “all-operations”, “read”, or “write”.
  buckets = ["my-tf-bucket"]
}
```

For full provider documentation with details on all options available, see [docs](./docs/) folder.

## Tested By
* Sep 12, 2022: Alexander Chernin (alexander.chernin@seagate.com) on Windows 10.
