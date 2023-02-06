terraform {
  required_providers {
    lyvecloud = {
      source = "Seagate/lyvecloud"
      version = "1.0.0"
    }
  }
}

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