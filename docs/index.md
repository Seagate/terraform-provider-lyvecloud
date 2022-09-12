---
page_title: "Provider: Lyve Cloud"
description: Manage Lyve Cloud with Terraform.
---

# Lyve Cloud Provider

This is a terraform provider plugin for managing [Lyve Cloud](https://www.seagate.com/gb/en/services/cloud/storage/) S3 buckets, objects, permissions and service accounts.
This project is based on code samples from the official [AWS provider](https://github.com/hashicorp/terraform-provider-aws).

## Example Provider Configuration

To manage buckets and objects you need to set the S3 API credentials.
To manage permissions and service accounts you need to set the Account API.
You can set either the credentials of the S3 API, the credentials of the Account API, or both.

```terraform
provider "lyvecloud" {
  //s3 api
  region = ""
  access_key = ""
  secret_key = ""
  endpoint_url = ""

  //acount api
  client_id = ""
  client_secret = ""
}
```

## Authentication

The Lyve Cloud provider offers the following methods of providing credentials for
authentication, in this order, and explained below:

- Static API key
- Environment variables

### Static API Key

Static credentials can be provided by adding the following variables in-line in the
Lyve CLoud provider block:


```hcl
provider "lyvecloud" {
  //s3 api
  region = "..."
  access_key = "..."
  secret_key = "..."
  endpoint_url = "...""

  //acount api
  client_id = "..."
  client_secret = "..."
}
```

### Environment variables

You can provide your configuration via the environment variables representing your Lyve Cloud credentials:

```
$ export LYVECLOUD_REGION="<Lyve Cloud region>"
$ export LYVECLOUD_ACCESS_KEY="<Access Key to the Lyve Cloud API>"
$ export LYVECLOUD_SECRET_KEY="<Secret Key to the Lyve Cloud API>"
$ export LYVECLOUD_ENDPOINT="<Lyve Cloud Endpoint URL>"

$ export LYVECLOUD_CLIENT_ID="<Lyve Cloud Account API Client ID>"
$ export LYVECLOUD_CLIENT_SECRET="<Lyve Cloud Account API Client Secret>"
```

When using this method, you may omit the
lyvecloud `provider` block entirely:

```hcl
resource "lyvecloud_s3_bucket" "my_bucket" {
  # .....
}
```

## Argument Reference

The following arguments are supported in the `provider` block:

* `access_key` - (Optional) Lyve Cloud access key. Can also be set with the `LYVECLOUD_ACCESS_KEY_ID` environment variable. Must be set to manage S3 resources(buckets and objects). 

* `secret_key` - (Optional) Lyve Cloud secret key. Can also be set with the `LYVECLOUD_SECRET_ACCESS_KEY` environment variable. Must be set to manage S3 resources(buckets and objects).

* `region` - (Optional) Lyve Cloud region where the provider will operate. Can also be set with the `LYVECLOUD_REGION` environment variable. Must be set to manage S3 resources(buckets and objects).

* `endpoint_url` - (Optional) Lyve Cloud Endpoint URL. Can also be set with the `LYVECLOUD_ENDPOINT` environment variable. Must be set to manage S3 resources(buckets and objects).

* `client_id` - (Optional) Lyve Cloud Account API Client ID. Can also be set with the `LYVECLOUD_CLIENT_ID` environment variable. Must be set to manage Account API resources(permissions and service accounts).

* `client_secret` - (Optional) Lyve Cloud Account API Client Secret. Can also be set with the `LYVECLOUD_CLIENT_SECRET` environment variable. Must be set to manage Account API resources(permissions and service accounts).
