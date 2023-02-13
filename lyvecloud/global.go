package lyvecloud

import "github.com/aws/aws-sdk-go/service/s3"

type Client struct {
	S3Client         *s3.S3
	AccountAPIClient *AuthData
}

const (
	S3           = "s3"
	AccountAPI   = "acc"
	NoSuchTagSet = "NoSuchTagSet"

	// HTTP requests
	SlashSeparator = "/"
	Enabled        = "enabled"

	// URLs
	TokenUrl        = "https://api.lyvecloud.seagate.com/v2/auth/token"
	PermissionUrl   = "https://api.lyvecloud.seagate.com/v2/permissions"
	SAUrl           = "https://api.lyvecloud.seagate.com/v2/service-accounts"
	UsageMonthlyUrl = "https://api.lyvecloud.seagate.com/v2/usage/monthly"
	UsageCurrentUrl = "https://api.lyvecloud.seagate.com/v2/usage/current"

	// headers
	Accept            = "Accept"
	Bearer            = "Bearer "
	ContentType       = "Content-Type"
	Authorization     = "Authorization"
	Json              = "application/json"
	UserAgent         = "User-Agent"
	TerraformProvider = "TerraformProvider/0.2.0"
)
