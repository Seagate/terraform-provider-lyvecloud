package lyvecloud

import "github.com/aws/aws-sdk-go/service/s3"

type Client struct {
	S3Client       *s3.S3
	AccAPIV1Client *AuthData
	AccAPIV2Client *AuthDataV2
}

const (
	S3           = "s3"
	AccountAPIV1 = "acc"
	AccountAPIV2 = "accv2"
	NoSuchTagSet = "NoSuchTagSet"
	AccessDenied = "AccessDenied"

	// HTTP requests
	SlashSeparator = "/"
	Enabled        = "enabled"

	// Account API V1 Auth
	AudienceUrl       = "https://lyvecloud/customer/api"
	ClientCredentials = "client_credentials"

	// URLs
	TokenUrl           = "https://auth.lyve.seagate.com/oauth/token"
	TokenUrlV2         = "https://api.lyvecloud.seagate.com/v2/auth/token"
	TokenUrlV2STG      = "https://api.us-west-1-stg.lyvecloud.seagate.com/v2/auth/token"
	PermissionUrl      = "https://api.lyvecloud.seagate.com/v1/permission"
	PermissionUrlV2    = "https://api.lyvecloud.seagate.com/v2/permissions/"
	PermissionUrlV2STG = "https://api.us-west-1-stg.lyvecloud.seagate.com/v2/permissions"
	SAUrl              = "https://api.lyvecloud.seagate.com/v1/service-account"
	SAUrlV2            = "https://api.lyvecloud.seagate.com/v2/service-accounts/"
	SAUrlV2STG         = "https://api.us-west-1-stg.lyvecloud.seagate.com/v2/service-accounts"

	// headers
	Accept        = "Accept"
	Bearer        = "Bearer "
	ContentType   = "Content-Type"
	Authorization = "Authorization"
	Json          = "application/json"
)
