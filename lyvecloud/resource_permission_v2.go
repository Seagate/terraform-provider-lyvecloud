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
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"name_prefix"},
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"name"},
			},
			"description": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true, // computed based on the chosen argument. all_buckets/prefix/buckets
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
				ConflictsWith: []string{"buckets", "bucket_prefix"},
			},
			"bucket_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"buckets", "all_buckets"},
			},
			"buckets": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ConflictsWith: []string{"bucket_prefix", "all_buckets"},
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
	if CheckCredentials(AccountAPIV2, meta.(Client)) {
		return fmt.Errorf("credentials for account api v2 are missing")
	}

	conn := *meta.(Client).AccAPIV2Client

	name := NameWithSuffix(d.Get("name").(string), d.Get("name_prefix").(string))
	description := d.Get("description").(string)
	actions := d.Get("actions").(string)

	var buckets []string
	prefix := d.Get("bucket_prefix").(string)
	bucketsList := d.Get("buckets").([]interface{})

	var permissionType string

	if d.Get("all_buckets").(bool) {
		permissionType = "all-buckets"
	} else if prefix != "" {
		permissionType = "bucket-prefix"
	} else if bucketsList != nil {
		permissionType = "bucket-names"
		var err error
		buckets, err = convertBucketsList(bucketsList)
		if err != nil {
			return fmt.Errorf("error creating permission: %w", err)
		}
	} else {
		return fmt.Errorf("one of the following keys must be used: buckets/bucket_prefix/all_buckets")
	}

	// create input for CreatePermissionV2
	createPermissionInput := PermissionV2{
		Name:        name,
		Description: description,
		Type:        permissionType,
		Actions:     actions,
		Prefix:      prefix,
		Buckets:     buckets,
	}

	resp, err := conn.CreatePermissionV2(&createPermissionInput)
	if err != nil {
		return fmt.Errorf("error creating permission: %w", err)
	}
	d.SetId(resp.ID)

	return resourcePermissionV2Read(d, meta)
}

func resourcePermissionV2Read(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(AccountAPIV2, meta.(Client)) {
		return fmt.Errorf("credentials for account api v2 are missing")
	}

	conn := *meta.(Client).AccAPIV2Client

	permissionId := d.Id()

	resp, err := conn.GetPermissionV2(permissionId)
	if err != nil {
		return fmt.Errorf("error reading permission (%s): %w", permissionId, err)
	}

	d.Set("id", resp.Id)
	d.Set("name", resp.Name)
	d.Set("description", resp.Description)
	d.Set("type", resp.Type)
	d.Set("actions", resp.Actions)

	if resp.Type != "all-buckets" {
		d.Set("bucket_prefix", resp.Prefix)
		d.Set("buckets", resp.Buckets)
	}

	d.Set("ready_state", resp.ReadyState)

	return nil
}

func resourcePermissionV2Update(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(AccountAPIV2, meta.(Client)) {
		return fmt.Errorf("credentials for account api v2 are missing")
	}

	conn := *meta.(Client).AccAPIV2Client

	permissionId := d.Id()

	name := NameWithSuffix(d.Get("name").(string), d.Get("name_prefix").(string))
	description := d.Get("description").(string)
	actions := d.Get("actions").(string)

	var buckets []string
	prefix := d.Get("bucket_prefix").(string)
	bucketsList := d.Get("buckets").([]interface{})

	var permissionType string

	if d.Get("all_buckets").(bool) {
		permissionType = "all-buckets"
		buckets = append(buckets, "*")
	} else if prefix != "" {
		permissionType = "bucket-prefix"
	} else if bucketsList != nil {
		permissionType = "bucket-names"
		var err error
		buckets, err = convertBucketsList(bucketsList)
		if err != nil {
			return fmt.Errorf("error updating permission: %w", err)
		}
	} else {
		return fmt.Errorf("one of the following keys must be used: buckets/bucket_prefix/all_buckets")
	}

	// update input for UpdatePermissionV2
	updatePermissinInput := PermissionV2{
		Name:        name,
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
	if CheckCredentials(AccountAPIV2, meta.(Client)) {
		return fmt.Errorf("credentials for account api v2 are missing")
	}

	conn := *meta.(Client).AccAPIV2Client

	_, err := conn.DeletePermissionV2(d.Id())
	if err != nil {
		return fmt.Errorf("error deleting permission: %w", err)
	}

	return nil
}
