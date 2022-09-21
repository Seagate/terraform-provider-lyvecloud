package lyvecloud

import "github.com/aws/aws-sdk-go/service/s3"

type Client struct {
	S3Client     *s3.S3
	AccApiClient *AuthData
}

type AuthData struct {
	Access_token string
	Expires_in   int
	Token_type   string
}

const (
	SlashSeparator = "/"
	Bearer         = "Bearer "

	Put           = "PUT"
	Post          = "POST"
	Delete        = "DELETE"
	AccessDenied  = "AccessDenied"
	ContentType   = "Content-Type"
	NoSuchTagSet  = "NoSuchTagSet"
	Authorization = "Authorization"
	Json          = "application/json"

	TokenUrl      = "https://auth.lyve.seagate.com/oauth/token"
	PermissionUrl = "https://api.lyvecloud.seagate.com/v1/permission"
	SAUrl         = "https://api.lyvecloud.seagate.com/v1/service-account"
	ClientReq     = `{"client_id":"%s", "client_secret":"%s", "audience":"https://lyvecloud/customer/api", "grant_type":"client_credentials"}`

	Account = "acc"
	S3      = "s3"

	UnauthorizedMessage = "unauthorized: incorrect client_id or client_secret values"
)
