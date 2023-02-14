---
page_title: "Provider: Lyve Cloud"
description: Manage Lyve Cloud with Terraform.
---

# Lyve Cloud Provider

This is a terraform provider plugin for managing [Lyve Cloud](https://www.seagate.com/gb/en/services/cloud/storage/) S3 buckets, objects, permissions and service accounts.
This project is based on the official [AWS provider](https://github.com/hashicorp/terraform-provider-aws).

## Example Provider Configuration

To manage buckets and objects, the S3 API credentials must be set within the `s3` block. To manage Permissions and Service Account, credentials for Account API should be set in the `account` block. Both of the mentioned blocks is optional and is only required if there is utilization of a resource that depends on it.

-> To generate credentials for Account API, see the [following document](https://help.lyvecloud.seagate.com/en/using-account-api.html#generating-account-api-credentials).

```terraform
provider "lyvecloud" {
  s3 {
    region = ""
    access_key = ""
    secret_key = ""
    endpoint_url = ""
  }

  account {
    account_id = ""
    access_key = ""
    secret = ""
  }
}
```

## Authentication

The Lyve Cloud provider offers the following methods of providing credentials for
authentication, in this order, and explained below:

- Static API key
- Environment variables

### Static API Key

Static credentials can be provided in the
Lyve Cloud provider block by including the following blocks, which contains variables in-line:

```hcl
provider "lyvecloud" {
  s3 {
    region = "..."
    access_key = "..."
    secret_key = "..."
    endpoint_url = "..."
  }

  account {
    account_id = "..."
    access_key = "..."
    secret = "..."
  }
}
```

### Environment variables

You can provide your configuration via the environment variables representing your Lyve Cloud credentials:

```
$ export LYVECLOUD_S3_REGION="<Lyve Cloud region>"
$ export LYVECLOUD_S3_ACCESS_KEY="<Access Key to the Lyve Cloud API>"
$ export LYVECLOUD_S3_SECRET_KEY="<Secret Key to the Lyve Cloud API>"
$ export LYVECLOUD_S3_ENDPOINT="<Lyve Cloud Endpoint URL>"

$ export LYVECLOUD_ACCOUNT_ID="<Lyve Cloud Account API Client Account ID>"
$ export LYVECLOUD_ACCOUNT_ACCESS_KEY="<Lyve Cloud Account API Client Access Key>"
$ export LYVECLOUD_ACCOUNT_SECRET="<Lyve Cloud Account API Client Secret>"

```

~> When using environment variables, an empty block for each type of API is required to allow provider configurations from environment variables to be specified.

```terraform
provider "lyvecloud" {
  s3 {} # When using S3 API credentials environment variables
}
```

## Argument Reference

The following arguments are supported in the `provider` block:

* `s3` - (Optional) Configuration block to use S3 API credentials.
  * `access_key` - (Required) Lyve Cloud access key. Can also be set with the `LYVECLOUD_S3_ACCESS_KEY` environment variable. Must be set to manage S3 resources(buckets and objects). 
  * `secret_key` - (Required) Lyve Cloud secret key. Can also be set with the `LYVECLOUD_S3_SECRET_KEY` environment variable. Must be set to manage S3 resources(buckets and objects).
  * `region` - (Required) Lyve Cloud region where the provider will operate. Can also be set with the `LYVECLOUD_S3_REGION` environment variable. Must be set to manage S3 resources(buckets and objects).
  * `endpoint_url` - (Required) Lyve Cloud Endpoint URL. Can also be set with the `LYVECLOUD_S3_ENDPOINT` environment variable. Must be set to manage S3 resources(buckets and objects).

* `account` - (Optional) Configuration block to use Account API credentials.
  * `account_id` - (Required) Lyve Cloud Account API Client Account ID. Can also be set with the `LYVECLOUD_ACCOUNT_ID` environment variable. Must be set to manage Account API resources.
  * `access_key` - (Required) Lyve Cloud Account API Client Access Key. Can also be set with the `LYVECLOUD_ACCOUNT_ACCESS_KEY` environment variable. Must be set to manage Account API resources.
  * `secret` - (Required) Lyve Cloud Account API Client Secret. Can also be set with the `LYVECLOUD_ACCOUNT_SECRET` environment variable. Must be set to manage Account API resources(permissions and service accounts).