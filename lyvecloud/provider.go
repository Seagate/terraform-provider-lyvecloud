package lyvecloud

import (
	"context"
	"errors"

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
				Description: "The access key for API operations.",
				DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_ACCESS_KEY", nil),
			},
			"secret_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The secret key for API operations.",
				DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_SECRET_KEY", nil),
			},
			"region": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Lyve Cloud region.",
				DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_REGION", nil),
			},
			"endpoint_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Lyve Cloud endpoint URL",
				DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_ENDPOINT", nil),
			},
			"client_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The client ID for Account API operations.",
				DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_CLIENT_ID", nil),
			},
			"client_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The region where Lyve Cloud operations will take place.",
				DefaultFunc: schema.EnvDefaultFunc("LYVECLOUD_CLIENT_SECRET", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"lyvecloud_s3_bucket":                           ResourceBucket(),
			"lyvecloud_s3_object":                           ResourceObject(),
			"lyvecloud_s3_object_copy":                      ResourceObjectCopy(),
			"lyvecloud_permission":                          ResourcePermission(),
			"lyvecloud_permission_v2":                       ResourcePermission(),
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

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var (
		client           *s3.S3
		s3Client         bool
		accountAPIClient bool
		accApiClient     *AuthData
	)

	region := d.Get("region").(string)
	clientId := d.Get("client_id").(string)
	accessKey := d.Get("access_key").(string)
	secretKey := d.Get("secret_key").(string)
	endpointUrl := d.Get("endpoint_url").(string)
	clientSecret := d.Get("client_secret").(string)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if (accessKey != "") && (secretKey != "") && (region != "") && (endpointUrl != "") {
		s3Client = true
		// Setting configs
		s3Config := &aws.Config{
			Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
			Endpoint:         aws.String(endpointUrl),
			Region:           aws.String(region),
			DisableSSL:       aws.Bool(true),
			S3ForcePathStyle: aws.Bool(true),
		}

		// Create S3 service session
		sess, err := session.NewSession(s3Config)

		// Create S3 service client
		client = s3.New(sess)

		if err != nil {
			return nil, diag.FromErr(errors.New("error creating aws sdk client"))
		}
	}

	if (clientId != "") && (clientSecret != "") {
		accountAPIClient = true
		// create account api client config
		credentials := Auth{
			ClientID:     clientId,
			ClientSecret: clientSecret,
		}

		// create account api client
		var err error
		accApiClient, err = AuthAccountAPI(&credentials)
		if err != nil {
			return nil, diag.FromErr(err)
		}
	}

	if s3Client && accountAPIClient {
		return Client{S3Client: client, AccApiClient: accApiClient}, diags
	} else if s3Client {
		return Client{S3Client: client}, diags
	} else if accountAPIClient {
		return Client{AccApiClient: accApiClient}, diags
	}

	// Create S3 service session
	sess, err := session.NewSession(nil)

	// Create S3 service client
	client = s3.New(sess)

	if err != nil {
		return nil, diag.FromErr(err)
	}

	// create account api client
	credentials := Auth{}
	accApiClient, err = AuthAccountAPI(&credentials)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return Client{S3Client: client, AccApiClient: accApiClient}, diags
}
