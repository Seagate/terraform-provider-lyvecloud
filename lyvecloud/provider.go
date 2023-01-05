package lyvecloud

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"access_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The access key for S3 API operations.",
				DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_S3_ACCESS_KEY", nil),
			},
			"secret_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The secret key for S3 API operations.",
				DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_S3_SECRET_KEY", nil),
			},
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Lyve Cloud region for S3 API operations.",
				DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_S3_REGION", nil),
			},
			"endpoint_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Lyve Cloud endpoint URL for S3 API operations.",
				DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_S3_ENDPOINT", nil),
			},
			"client_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The client ID for Account API V1 operations.",
				DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_AAPIV1_CLIENT_ID", nil),
			},
			"client_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The client secret for Account API V1 operations.",
				DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_AAPIV1_CLIENT_SECRET", nil),
			},
			"accountId": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Unique identifier of the Lyve Cloud Account API v2.",
				DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_AAPIV2_ACCOUNT_ID", nil),
			},
			"accessKey": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The access key is generated when you generate Account API v2 credentails.",
				DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_AAPIV2_ACCESS_KEY", nil),
			},
			"secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The secret key is generated when you generate Account API v2 credentials.",
				DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_AAPIV2_SECRET", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"lyvecloud_s3_bucket":                           ResourceBucket(),
			"lyvecloud_s3_object":                           ResourceObject(),
			"lyvecloud_s3_object_copy":                      ResourceObjectCopy(),
			"lyvecloud_permission":                          ResourcePermission(),
			"lyvecloud_permission_v2":                       ResourcePermissionV2(),
			"lyvecloud_service_account":                     ResourceServiceAccount(),
			"lyvecloud_service_account_v2":                  ResourceServiceAccountV2(),
			"lyvecloud_s3_bucket_object_lock_configuration": ResourceBucketObjectLockConfiguration(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"lyvecloud_s3_bucket": DataSourceBucket(),
			"lyvecloud_s3_object": DataSourceObject(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

// createS3Client creates AWS SDK client.
func createS3Client(region, accessKey, secretKey, endpointUrl string) (*s3.S3, error) {
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         aws.String(endpointUrl),
		Region:           aws.String(region),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	}
	sess, err := session.NewSession(s3Config)
	if err != nil {
		return nil, err
	}
	client := s3.New(sess)
	return client, nil
}

// createAccApiClient creates Account API v1 client.
func createAccAPIV1Client(clientId, clientSecret string) (*AuthData, error) {
	credentials := Auth{
		ClientID:     clientId,
		ClientSecret: clientSecret,
	}
	accApiClient, err := AuthAccountAPI(&credentials)
	if err != nil {
		return nil, fmt.Errorf("error authenticating account APIv1: %w", err)
	}
	return accApiClient, nil
}

// createAccApiClient creates Account API v1 client.
func createAccAPIV2Client(accountId, accessKey, secret string) (*AuthDataV2, error) {
	credentials := AuthV2{
		AccountID: accountId,
		AccessKey: accessKey,
		Secret:    secret,
	}
	accAPIV2Client, err := AuthAccountAPIV2(&credentials)
	if err != nil {
		return nil, fmt.Errorf("error authenticating account APIv2: %w", err)
	}

	return accAPIV2Client, nil
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var (
		s3Client         *s3.S3
		accAPIV1Client   *AuthData
		accAPIV2Client   *AuthDataV2
		s3Bool           bool
		accountAPIV1Bool bool
		accountAPIV2Bool bool
	)

	// S3 API
	region := d.Get("region").(string)
	accessKey := d.Get("access_key").(string)
	secretKey := d.Get("secret_key").(string)
	endpointUrl := d.Get("endpoint_url").(string)

	// AAPIV1
	clientId := d.Get("client_id").(string)
	clientSecret := d.Get("client_secret").(string)

	// AAPIV2
	accountId := d.Get("accountId").(string)
	accessKey := d.Get("accessKey").(string)
	secret := d.Get("accessKey").(string)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if (accessKey != "") && (secretKey != "") && (region != "") && (endpointUrl != "") {
		s3Bool = true
		var err error

		s3Client, err = createS3Client(region, accessKey, secretKey, endpointUrl)
		if err != nil {
			return nil, diag.FromErr(err)
		}
	}

	if (clientId != "") && (clientSecret != "") {
		accountAPIV1Bool = true
		var err error
		accAPIV1Client, err = createAccAPIV1Client(clientId, clientSecret)
		if err != nil {
			return nil, diag.FromErr(err)
		}
	}

	if s3Bool && accountAPIV1Bool {
		return Client{S3Client: s3Client, AccApiClient: accAPIV1Client}, diags
	} else if s3Bool {
		return Client{S3Client: s3Client}, diags
	} else if accountAPIV1Bool {
		return Client{AccApiClient: accAPIV1Client}, diags
	}
	return nil, diag.FromErr(errors.New("no valid credentials provided"))
}
