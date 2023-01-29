package lyvecloud

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceServiceAccountV2() *schema.Resource {

	return &schema.Resource{
		Create: resourceServiceAccountV2Create,
		Read:   resourceServiceAccountV2Read,
		Update: resourceServiceAccountV2Update,
		Delete: resourceServiceAccountV2Delete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
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
			"secret": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ready_state": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceServiceAccountV2Create(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(AccountAPIV2, meta.(Client)) {
		return fmt.Errorf("credentials for account api v2 are missing")
	}

	conn := *meta.(Client).AccAPIV2Client

	name := d.Get("name").(string)
	description := d.Get("description").(string)
	permissionsList := d.Get("permissions").([]interface{})

	permissions := []string{}
	for _, v := range permissionsList {
		permissions = append(permissions, v.(string))
	}

	serviceAccountInput := ServiceAccount{
		Name:        name,
		Description: description,
		Permissions: permissions,
	}

	resp, err := conn.CreateServiceAccountV2(&serviceAccountInput)
	if err != nil {
		return fmt.Errorf("error creating service account: %w", err)
	}

	d.SetId(resp.ID)

	d.Set("access_key", resp.Accesskey)
	d.Set("secret", resp.Secret)

	return resourceServiceAccountV2Read(d, meta)
}

func resourceServiceAccountV2Read(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(AccountAPIV2, meta.(Client)) {
		return fmt.Errorf("credentials for account api v2 are missing")
	}

	conn := *meta.(Client).AccAPIV2Client

	serviceAccountId := d.Id()

	resp, err := conn.GetServiceAccountV2(serviceAccountId)
	if err != nil {
		return fmt.Errorf("error reading service account (%s): %w", serviceAccountId, err)
	}

	d.Set("id", resp.Id)
	d.Set("name", resp.Name)
	d.Set("description", resp.Description)
	d.Set("ready_state", resp.ReadyState)
	d.Set("permissions", resp.Permissions)
	d.Set("enabled", resp.Enabled)

	return nil
}

func resourceServiceAccountV2Update(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(AccountAPIV2, meta.(Client)) {
		return fmt.Errorf("credentials for account api v2 are missing")
	}

	conn := *meta.(Client).AccAPIV2Client

	serviceAccountId := d.Id()

	name := d.Get("name").(string)
	description := d.Get("description").(string)
	permissionsList := d.Get("permissions").([]interface{})

	permissions := []string{}
	for _, v := range permissionsList {
		permissions = append(permissions, v.(string))
	}

	updateServiceAccountInput := ServiceAccount{
		Name:        name,
		Description: description,
		Permissions: permissions,
	}

	_, err := conn.UpdateServiceAccountV2(serviceAccountId, &updateServiceAccountInput)
	if err != nil {
		return fmt.Errorf("error updating service account: %w", err)
	}

	return resourceServiceAccountV2Read(d, meta)
}

func resourceServiceAccountV2Delete(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(AccountAPIV2, meta.(Client)) {
		return fmt.Errorf("credentials for account api v2 are missing")
	}

	conn := *meta.(Client).AccAPIV2Client

	_, err := conn.DeleteServiceAccountV2(d.Id())
	if err != nil {
		return fmt.Errorf("error deleting service account: %w", err)
	}

	return nil
}
