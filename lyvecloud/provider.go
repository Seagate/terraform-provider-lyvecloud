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
			"account": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "The credentials for the Account API are used to manage permissions and service accounts.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unique identifier of the Lyve Cloud Account API.",
							DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_ACCOUNT_ID", nil),
						},
						"access_key": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The access key is generated when you generate Account API credentails.",
							DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_ACCOUNT_ACCESS_KEY", nil),
						},
						"secret": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The secret key is generated when you generate Account API credentials.",
							DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_ACCOUNT_SECRET", nil),
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
			"lyvecloud_service_account":                     ResourceServiceAccount(),
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

// createAccAPIClient creates Account API v2 client.
func createAccountAPIClient(accountId, accessKey, secret string) (*AuthData, error) {
	credentials := AuthRequest{
		AccountID: accountId,
		AccessKey: accessKey,
		Secret:    secret,
	}
	accountAPIClient, err := AuthAccountAPI(&credentials)
	if err != nil {
		return nil, fmt.Errorf("error authenticating account API: %w", err)
	}

	return accountAPIClient, nil
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var s3Client *s3.S3
	var accountAPIClient *AuthData
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

	if accountAPI, ok := d.Get("account").([]interface{}); ok && len(accountAPI) > 0 && accountAPI[0] != nil {
		accountAPIAttr := accountAPI[0].(map[string]interface{})

		var accountId, accessKey, secret string

		if v, ok := accountAPIAttr["account_id"].(string); ok && v != "" {
			accountId = v
		} else {
			return nil, diag.FromErr(errors.New("account_id must be set and contain a non-empty value"))
		}

		if v, ok := accountAPIAttr["access_key"].(string); ok && v != "" {
			accessKey = v
		} else {
			return nil, diag.FromErr(errors.New("access_key must be set and contain a non-empty value"))
		}

		if v, ok := accountAPIAttr["secret"].(string); ok && v != "" {
			secret = v
		} else {
			return nil, diag.FromErr(errors.New("secret must be set and contain a non-empty value"))
		}

		accountAPIClient, err = createAccountAPIClient(accountId, accessKey, secret)
		if err != nil {
			return nil, diag.FromErr(err)
		}
	}

	return Client{S3Client: s3Client, AccountAPIClient: accountAPIClient}, nil
}
