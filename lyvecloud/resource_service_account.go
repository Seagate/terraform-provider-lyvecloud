package lyvecloud

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceServiceAccount() *schema.Resource {

	return &schema.Resource{
		Create: resourceServiceAccountCreate,
		Read:   resourceServiceAccountRead,
		Update: resourceServiceAccountUpdate,
		Delete: resourceServiceAccountDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 128),
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

func resourceServiceAccountCreate(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(AccountAPI, meta.(Client)) {
		return fmt.Errorf("credentials for account api are missing")
	}

	conn := *meta.(Client).AccountAPIClient

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

	resp, err := conn.CreateServiceAccount(&serviceAccountInput)
	if err != nil {
		return fmt.Errorf("error creating service account: %w", err)
	}

	d.SetId(resp.ID)

	d.Set("access_key", resp.Accesskey)
	d.Set("secret", resp.Secret)

	return resourceServiceAccountRead(d, meta)
}

func resourceServiceAccountRead(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(AccountAPI, meta.(Client)) {
		return fmt.Errorf("credentials for account api are missing")
	}

	conn := *meta.(Client).AccountAPIClient

	serviceAccountId := d.Id()

	resp, err := conn.GetServiceAccount(serviceAccountId)

	if !d.IsNewResource() && err.Error() == ServiceAccountNotFound {
		log.Printf("[WARN] Service Account (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

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

func resourceServiceAccountUpdate(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(AccountAPI, meta.(Client)) {
		return fmt.Errorf("credentials for account api are missing")
	}

	conn := *meta.(Client).AccountAPIClient

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

	_, err := conn.UpdateServiceAccount(serviceAccountId, &updateServiceAccountInput)

	if !d.IsNewResource() && err.Error() == ServiceAccountNotFound {
		log.Printf("[WARN] Service Account (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error updating service account: %w", err)
	}

	return resourceServiceAccountRead(d, meta)
}

func resourceServiceAccountDelete(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(AccountAPI, meta.(Client)) {
		return fmt.Errorf("credentials for account api are missing")
	}

	conn := *meta.(Client).AccountAPIClient

	_, err := conn.DeleteServiceAccount(d.Id())
	if err != nil {
		return fmt.Errorf("error deleting service account: %w", err)
	}

	return nil
}
