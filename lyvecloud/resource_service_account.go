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
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"permissions": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
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
		},
	}
}

func resourceServiceAccountCreate(d *schema.ResourceData, m interface{}) error {
	if CheckCredentials(Account, m.(Client)) {
		return fmt.Errorf("credentials for account API(client_id, client_secret) are missing")
	}

	c := m.(Client).AccApiClient

	service_account := d.Get("service_account").(string)
	description := d.Get("description").(string)
	permissionsList := d.Get("permissions").([]interface{})

	permissions := []string{}
	for _, v := range permissionsList {
		permissions = append(permissions, v.(string))
	}

	resp, err := c.CreateServiceAccount(service_account, description, permissions)
	if err != nil {
		return err
	}

	d.SetId(resp.ID)

	d.Set("access_key", resp.Access_key)
	d.Set("access_secret", resp.Access_Secret)
	return resourceServiceAccountRead(d, m)
}

func resourceServiceAccountRead(d *schema.ResourceData, m interface{}) error {
	d.Set("id", d.Id())
	return nil
}

func resourceServiceAccountUpdate(d *schema.ResourceData, m interface{}) error {
	// currently useless
	return nil
}

func resourceServiceAccountDelete(d *schema.ResourceData, m interface{}) error {
	if CheckCredentials(Account, m.(Client)) {
		return fmt.Errorf("credentials for account API(client_id, client_secret) are missing")
	}

	c := m.(Client).AccApiClient

	_, err := c.DeleteServiceAccount(d.Id())
	if err != nil {
		return fmt.Errorf("error deleting service account")
	}

	return nil
}
