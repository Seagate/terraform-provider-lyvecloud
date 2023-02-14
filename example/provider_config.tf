terraform {
  required_providers {
    lyvecloud = {
      source = "Seagate/lyvecloud"
      version = "0.2.0"
    }
  }
}

provider "lyvecloud" {
  // client credentials for managaing s3 resources
  s3 {
    region = "..."
    access_key = "..."
    secret_key = "..."
    endpoint_url = "..."
  }

  // client credentials for account api to manage permissions and service accounts
  account {
    account_id = "..."
    access_key = "..."
    secret = "..."
  }
}