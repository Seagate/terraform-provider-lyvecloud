package lyvecloud

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourcePermission() *schema.Resource {

	return &schema.Resource{
		Create: resourceCreatePermission,
		Read:   resourceReadPermission,
		Update: resourceUpdatePermission,
		Delete: resourceDeletePermission,

		Schema: map[string]*schema.Schema{
			"permission": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"actions": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"buckets": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceCreatePermission(d *schema.ResourceData, m interface{}) error {
	if CheckCredentials(Account, m.(Client)) {
		return fmt.Errorf("credentials for account API(client_id, client_secret) are missing")
	}

	c := m.(Client).AccApiClient

	permission := d.Get("permission").(string)
	description := d.Get("description").(string)
	actions := d.Get("actions").(string)
	bucketsList := d.Get("buckets").([]interface{})

	buckets := []string{}
	for _, v := range bucketsList {
		buckets = append(buckets, v.(string))
	}

	resp, err := c.CreatePermission(permission, description, actions, buckets)
	if err != nil {
		return fmt.Errorf("error creating Permission: %w", err)
	}
	d.SetId(resp.ID)

	return resourceReadPermission(d, m)
}

func resourceReadPermission(d *schema.ResourceData, m interface{}) error {
	d.Set("id", d.Id())
	return nil
}

func resourceUpdatePermission(d *schema.ResourceData, m interface{}) error {
	// currently useless
	return nil
}

func resourceDeletePermission(d *schema.ResourceData, m interface{}) error {
	if CheckCredentials(Account, m.(Client)) {
		return fmt.Errorf("credentials for account API(client_id, client_secret) are missing")
	}

	c := m.(Client).AccApiClient

	_, err := c.DeletePermission(d.Id())
	if err != nil {
		return fmt.Errorf("error deleting permission")
	}

	return nil
}
