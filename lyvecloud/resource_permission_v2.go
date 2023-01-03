package lyvecloud

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourcePermissionV2() *schema.Resource {

	return &schema.Resource{
		Create: resourcePermissionV2Create,
		Read:   resourcePermissionV2Read,
		Update: resourcePermissionV2Update,
		Delete: resourcePermissionV2Delete,

		Schema: map[string]*schema.Schema{
			"permission": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true, // computed based on the chosen argument. all_buckets/bucket_prefix/bucket_names
			},
			"actions": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"all-operations",
					"read-only",
					"write-only",
				}, false),
			},
			"all_buckets": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"buckets", "bucket-prefix"},
			},
			"bucket_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"buckets", "all-buckets"},
			},
			"bucket_names": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"ready_state": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourcePermissionV2Create(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(Account, meta.(Client)) {
		return fmt.Errorf("credentials for account api(client_id, client_secret) are missing")
	}

	conn := *meta.(Client).AccApiClientV2

	permission := d.Get("permission").(string)
	description := d.Get("description").(string)
	actions := d.Get("actions").(string)
	prefix := d.Get("bucket_prefix").(string)

	var permissionType string
	buckets := []string{}

	if _, ok := d.GetOk("all_buckets"); ok {
		permissionType = "all-buckets"
	} else if v, ok := d.GetOk("bucket_prefix"); ok {
		permissionType = "bucket-prefix"
		buckets = append(buckets, v.(string))
	} else if _, ok := d.GetOk("bucket_names"); ok {
		permissionType = "bucket-names"
		bucketsList := d.Get("buckets").([]interface{})
		for _, v := range bucketsList {
			buckets = append(buckets, v.(string))
		}
	}

	// create input for CreatePermissionV2
	createPermissinInput := PermissionV2{
		Name:        permission,
		Description: description,
		Type:        permissionType,
		Actions:     actions,
		Prefix:      prefix,
		Buckets:     buckets,
	}

	resp, err := conn.CreatePermissionV2(&createPermissinInput)
	if err != nil {
		return fmt.Errorf("error creating permission: %w", err)
	}
	d.SetId(resp.ID)

	return resourcePermissionV2Read(d, meta)
}

func resourcePermissionV2Read(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(Account, meta.(Client)) {
		return fmt.Errorf("credentials for account api(client_id, client_secret) are missing")
	}

	conn := *meta.(Client).AccApiClientV2

	permissionId := d.Id()

	resp, err := conn.GetPermissionV2(permissionId)
	if err != nil {
		return fmt.Errorf("error reading permission: %w", err)
		// try to remove it if error?
	}

	d.Set("id", resp.Id)
	d.Set("permission", resp.Name)
	d.Set("description", resp.Description)
	d.Set("type", resp.Type)
	d.Set("actions", resp.Actions)
	d.Set("bucket_prefix", resp.Prefix)
	d.Set("bucket_names", resp.Buckets)
	d.Set("ready_state", resp.ReadyState)

	return nil
}

func resourcePermissionV2Update(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(Account, meta.(Client)) {
		return fmt.Errorf("credentials for account api(client_id, client_secret) are missing")
	}

	conn := *meta.(Client).AccApiClientV2

	permissionId := d.Id()

	permission := d.Get("permission").(string)
	description := d.Get("description").(string)
	actions := d.Get("actions").(string)
	prefix := d.Get("bucket_prefix").(string)

	var permissionType string
	buckets := []string{}

	if _, ok := d.GetOk("all_buckets"); ok {
		permissionType = "all-buckets"
	} else if v, ok := d.GetOk("bucket_prefix"); ok {
		permissionType = "bucket-prefix"
		buckets = append(buckets, v.(string))
	} else if _, ok := d.GetOk("bucket_names"); ok {
		permissionType = "bucket-names"
		bucketsList := d.Get("buckets").([]interface{})
		for _, v := range bucketsList {
			buckets = append(buckets, v.(string))
		}
	}

	updatePermissinInput := PermissionV2{
		Name:        permission,
		Description: description,
		Type:        permissionType,
		Actions:     actions,
		Prefix:      prefix,
		Buckets:     buckets,
	}

	_, err := conn.UpdatePermissionV2(permissionId, &updatePermissinInput)
	if err != nil {
		return fmt.Errorf("error updating permission: %w", err)
	}

	return resourcePermissionV2Read(d, meta)
}

func resourcePermissionV2Delete(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(Account, meta.(Client)) {
		return fmt.Errorf("credentials for account api(client_id, client_secret) are missing")
	}

	conn := *meta.(Client).AccApiClientV2

	_, err := conn.DeletePermissionV2(d.Id())
	if err != nil {
		return fmt.Errorf("error deleting permission: %w", err)
	}

	return nil
}
