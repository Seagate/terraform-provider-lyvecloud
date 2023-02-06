---
page_title: "Provider: Lyve Cloud"
description: Manage Lyve Cloud with Terraform.
---

# Lyve Cloud Provider

This is a terraform provider plugin for managing [Lyve Cloud](https://www.seagate.com/gb/en/services/cloud/storage/) S3 buckets, objects, permissions and service accounts.
This project is based on the official [AWS provider](https://github.com/hashicorp/terraform-provider-aws).

## Example Provider Configuration

To manage buckets and objects, the S3 API credentials must be set within the `s3` block. To manage Permissions and Service Account, credentials for Account API v1 should be set in the `account_v1` block or for Account API v2 in the `account_v2` block. Each of the mentioned blocks is optional and is only required if there is utilization of a resource that depends on it.

-> To obtain the client_id and client_secret for Account API v1, a support ticket must be created requesting the credentials for the Account API v1.

-> To generate credentials for Account API v2, see the [following document](https://help.lyvecloud.seagate.com/en/using-account-api.html#generating-account-api-credentials).

```terraform
provider "lyvecloud" {
  s3 {
    region = ""
    access_key = ""
    secret_key = ""
    endpoint_url = ""
  }

  account_v1 {
    client_id = ""
    client_secret = ""
  }

  account_v2 {
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

  account_v1 {
    client_id = "..."
    client_secret = "..."
  }

  account_v2 {
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

$ export LYVECLOUD_AAPIV1_CLIENT_ID="<Lyve Cloud Account API Client v1 ID>"
$ export LYVECLOUD_AAPIV1_CLIENT_SECRET="<Lyve Cloud Account API v1 Client Secret>"

$ export LYVECLOUD_AAPIV2_ACCOUNT_ID="<Lyve Cloud Account API Client v2 Account ID>"
$ export LYVECLOUD_AAPIV2_ACCESS_KEY="<Lyve Cloud Account API Client v2 Access Key>"
$ export LYVECLOUD_AAPIV2_SECRET="<Lyve Cloud Account API Client v2 Secret>"

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

* `account_v1` - (Optional) Configuration block to use Account API v1 credentials.
  * `client_id` - (Required) Lyve Cloud Account API Client ID. Can also be set with the `LYVECLOUD_AAPIV1_CLIENT_ID` environment variable. Must be set to manage Account API v1 resources.
  * `client_secret` - (Required) Lyve Cloud Account API Client Secret. Can also be set with the `LYVECLOUD_AAPIV1_CLIENT_SECRET` environment variable. Must be set to manage Account API v1 resources.

* `account_v2` - (Optional) Configuration block to use Account API v2 credentials.
  * `account_id` - (Required) Lyve Cloud Account API Client v2 Account ID. Can also be set with the `LYVECLOUD_AAPIV2_ACCOUNT_ID` environment variable. Must be set to manage Account API v2 resources.
  * `access_key` - (Required) Lyve Cloud Account API Client v2 Access Key. Can also be set with the `LYVECLOUD_AAPIV2_ACCESS_KEY` environment variable. Must be set to manage Account API v2 resources.
  * `secret` - (Required) Lyve Cloud Account API Client v2 Secret. Can also be set with the `LYVECLOUD_AAPIV2_SECRET` environment variable. Must be set to manage Account API v2 resources(permissions and service accounts).
