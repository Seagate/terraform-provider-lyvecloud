package lyvecloud

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceServiceAccount() *schema.Resource {

	return &schema.Resource{
		Create: resourceServiceAccountCreate,
		Read:   resourceServiceAccountRead,
		Update: resourceServiceAccountUpdate,
		Delete: resourceServiceAccountDelete,

		Schema: map[string]*schema.Schema{
			"service_account": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true, // should be generated if left empty
			},
			"permissions": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"access_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"access_secret": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceServiceAccountCreate(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(AccountAPIV1, meta.(Client)) {
		return fmt.Errorf("credentials for account api are missing")
	}

	conn := *meta.(Client).AccAPIV1Client

	service_account := d.Get("service_account").(string)
	description := d.Get("description").(string)
	permissionsList := d.Get("permissions").([]interface{})

	permissions := []string{}
	for _, v := range permissionsList {
		permissions = append(permissions, v.(string))
	}

	serviceAccountInput := ServiceAccount{
		Name:        service_account,
		Description: description,
		Permissions: permissions,
	}

	resp, err := conn.CreateServiceAccount(&serviceAccountInput)
	if err != nil {
		return fmt.Errorf("error creating service account: %w", err)
	}

	d.SetId(resp.ID)

	d.Set("access_key", resp.Accesskey)
	d.Set("access_secret", resp.AccessSecret)
	return resourceServiceAccountRead(d, meta)
}

func resourceServiceAccountRead(d *schema.ResourceData, meta interface{}) error {
	d.Set("id", d.Id())
	return nil
}

func resourceServiceAccountUpdate(d *schema.ResourceData, meta interface{}) error {
	// not supported in account api v1
	return nil
}

func resourceServiceAccountDelete(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(AccountAPIV1, meta.(Client)) {
		return fmt.Errorf("credentials for account api are missing")
	}

	conn := *meta.(Client).AccAPIV1Client

	_, err := conn.DeleteServiceAccount(d.Id())
	if err != nil {
		return fmt.Errorf("error deleting service account: %w", err)
	}

	return nil
}
