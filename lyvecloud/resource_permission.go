package lyvecloud

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourcePermission() *schema.Resource {

	return &schema.Resource{
		Create: resourcePermissionCreate,
		Read:   resourcePermissionRead,
		Update: resourcePermissionUpdate,
		Delete: resourcePermissionDelete,

		Schema: map[string]*schema.Schema{
			"permission": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"actions": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"all-operations",
					"read-",
					"write",
				}, false),
			},
			"buckets": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourcePermissionCreate(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(Account, meta.(Client)) {
		return fmt.Errorf("credentials for account api(client_id, client_secret) are missing")
	}

	conn := *meta.(Client).AccApiClient

	permission := d.Get("permission").(string)
	description := d.Get("description").(string)
	actions := d.Get("actions").(string)
	bucketsList := d.Get("buckets").([]interface{})

	buckets := []string{}
	for _, v := range bucketsList {
		buckets = append(buckets, v.(string))
	}

	permissionInput := Permission{
		Name:        permission,
		Description: description,
		Actions:     actions,
		Buckets:     buckets,
	}

	resp, err := conn.CreatePermission(&permissionInput)
	if err != nil {
		return fmt.Errorf("error creating permission: %w", err)
	}
	d.SetId(resp.ID)

	return resourcePermissionRead(d, meta)
}

func resourcePermissionRead(d *schema.ResourceData, meta interface{}) error {
	d.Set("id", d.Id())
	return nil
}

func resourcePermissionUpdate(d *schema.ResourceData, meta interface{}) error {
	// currently useless
	return nil
}

func resourcePermissionDelete(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(Account, meta.(Client)) {
		return fmt.Errorf("credentials for account api(client_id, client_secret) are missing")
	}

	conn := *meta.(Client).AccApiClient

	_, err := conn.DeletePermission(d.Id())
	if err != nil {
		return fmt.Errorf("error deleting permission: %w", err)
	}

	return nil
}
