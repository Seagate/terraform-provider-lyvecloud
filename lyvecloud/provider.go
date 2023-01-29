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
			"s3": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "The credentials of the Lyve Cloud S3 API are used to manage buckets and objects.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_key": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The access key for S3 API operations.",
							DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_S3_ACCESS_KEY", nil),
						},
						"secret_key": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The secret key for S3 API operations.",
							DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_S3_SECRET_KEY", nil),
						},
						"region": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Lyve Cloud region for S3 API operations.",
							DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_S3_REGION", nil),
						},
						"endpoint_url": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Lyve Cloud endpoint URL for S3 API operations.",
							DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_S3_ENDPOINT", nil),
						},
					},
				},
			},
			"account_v1": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "The credentials for the Account API v1 are used to manage permissions and service accounts.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The client ID for Account API v1 operations.",
							DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_AAPIV1_CLIENT_ID", nil),
						},
						"client_secret": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The client secret for Account API v1 operations.",
							DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_AAPIV1_CLIENT_SECRET", nil),
						},
					},
				},
			},
			"account_v2": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "The credentials for the Account API v2 are used to manage permissions and service accounts.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unique identifier of the Lyve Cloud Account API v2.",
							DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_AAPIV2_ACCOUNT_ID", nil),
						},
						"access_key": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The access key is generated when you generate Account API v2 credentails.",
							DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_AAPIV2_ACCESS_KEY", nil),
						},
						"secret": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The secret key is generated when you generate Account API v2 credentials.",
							DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_AAPIV2_SECRET", nil),
						},
					},
				},
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

// createAccAPIV1Client creates Account API v1 client.
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

// createAccAPIV2Client creates Account API v1 client.
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
	var s3Client *s3.S3
	var accountAPIV1Client *AuthData
	var accountAPIV2Client *AuthDataV2
	var err error

	if s3, ok := d.Get("s3").([]interface{}); ok && len(s3) > 0 && s3[0] != nil {
		s3Attr := s3[0].(map[string]interface{})

		var region, accessKey, secretKey, endpointUrl string

		if v, ok := s3Attr["region"].(string); ok && v != "" {
			region = v
		} else {
			return nil, diag.FromErr(errors.New("region must be set and contain a non-empty value"))
		}

		if v, ok := s3Attr["access_key"].(string); ok && v != "" {
			accessKey = v
		} else {
			return nil, diag.FromErr(errors.New("access_key must be set and contain a non-empty value"))
		}

		if v, ok := s3Attr["secret_key"].(string); ok && v != "" {
			secretKey = v
		} else {
			return nil, diag.FromErr(errors.New("secret_key must be set and contain a non-empty value"))
		}

		if v, ok := s3Attr["endpoint_url"].(string); ok && v != "" {
			endpointUrl = v
		} else {
			return nil, diag.FromErr(errors.New("endpoint_url must be set and contain a non-empty value"))
		}

		s3Client, err = createS3Client(region, accessKey, secretKey, endpointUrl)
		if err != nil {
			return nil, diag.FromErr(err)
		}
	}

	if accountAPIV1, ok := d.Get("account_v1").([]interface{}); ok && len(accountAPIV1) > 0 && accountAPIV1[0] != nil {
		accountAPIV1Attr := accountAPIV1[0].(map[string]interface{})

		var clientId, clientSecret string

		if v, ok := accountAPIV1Attr["client_id"].(string); ok && v != "" {
			clientId = v
		} else {
			return nil, diag.FromErr(errors.New("client_id must be set and contain a non-empty value"))
		}

		if v, ok := accountAPIV1Attr["client_secret"].(string); ok && v != "" {
			clientSecret = v
		} else {
			return nil, diag.FromErr(errors.New("client_secret must be set and contain a non-empty value"))
		}

		accountAPIV1Client, err = createAccAPIV1Client(clientId, clientSecret)
		if err != nil {
			return nil, diag.FromErr(err)
		}
	}

	if accountAPIV2, ok := d.Get("account_v2").([]interface{}); ok && len(accountAPIV2) > 0 && accountAPIV2[0] != nil {
		accountAPIV2Attr := accountAPIV2[0].(map[string]interface{})

		var accountId, accessKey, secret string

		if v, ok := accountAPIV2Attr["account_id"].(string); ok && v != "" {
			accountId = v
		} else {
			return nil, diag.FromErr(errors.New("accountId must be set and contain a non-empty value"))
		}

		if v, ok := accountAPIV2Attr["access_key"].(string); ok && v != "" {
			accessKey = v
		} else {
			return nil, diag.FromErr(errors.New("accessKey must be set and contain a non-empty value"))
		}

		if v, ok := accountAPIV2Attr["secret"].(string); ok && v != "" {
			secret = v
		} else {
			return nil, diag.FromErr(errors.New("secret must be set and contain a non-empty value"))
		}

		accountAPIV2Client, err = createAccAPIV2Client(accountId, accessKey, secret)
		if err != nil {
			return nil, diag.FromErr(err)
		}
	}

	return Client{S3Client: s3Client, AccAPIV1Client: accountAPIV1Client, AccAPIV2Client: accountAPIV2Client}, nil
}
