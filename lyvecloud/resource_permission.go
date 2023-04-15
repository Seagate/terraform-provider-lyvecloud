package lyvecloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type EscapeError string

func ResourcePermission() *schema.Resource {

	return &schema.Resource{
		Create: resourcePermissionCreate,
		Read:   resourcePermissionRead,
		Update: resourcePermissionUpdate,
		Delete: resourcePermissionDelete,

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
				Computed: true, // computed based on the chosen argument. all_buckets/prefix/buckets/policy
			},
			"actions": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"all-operations",
					"read-only",
					"write-only",
				}, false),
				ConflictsWith: []string{"policy"},
			},
			"all_buckets": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"buckets", "bucket_prefix", "policy"},
				RequiredWith:  []string{"actions"},
			},
			"bucket_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"buckets", "all_buckets", "policy"},
				RequiredWith:  []string{"actions"},
			},
			"buckets": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ConflictsWith: []string{"bucket_prefix", "all_buckets", "policy"},
				RequiredWith:  []string{"actions"},
			},
			"policy": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"bucket_prefix", "all_buckets", "buckets", "actions"},
				StateFunc: func(v interface{}) string {
					json, _ := NormalizeJsonString(v)
					return json
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

func resourcePermissionCreate(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(AccountAPI, meta.(Client)) {
		return fmt.Errorf("credentials for account api are missing")
	}

	conn := *meta.(Client).AccountAPIClient

	name := NameWithSuffix(d.Get("name").(string), d.Get("name_prefix").(string))
	description := d.Get("description").(string)
	actions := d.Get("actions").(string)

	var policyJSON string
	var buckets []string
	prefix := d.Get("bucket_prefix").(string)

	var permissionType string

	if d.Get("all_buckets").(bool) {
		permissionType = "all-buckets"
	} else if prefix != "" {
		permissionType = "bucket-prefix"
	} else if v, ok := d.GetOk("buckets"); ok {
		permissionType = "bucket-names"
		var err error
		buckets, err = convertBucketsList(v.([]interface{}))
		if err != nil {
			return fmt.Errorf("error reading buckets list: %w", err)
		}
	} else if v, ok := d.GetOk("policy"); ok {
		permissionType = "policy"
		var err error
		policyJSON, err = NormalizeJsonString(v)
		if err != nil {
			return fmt.Errorf("policy (%s) is invalid JSON: %w", policyJSON, err)
		}
	} else {
		return fmt.Errorf("one of the following keys must be used: buckets/bucket_prefix/all_buckets/policy")
	}

	// create input for CreatePermission
	createPermissionInput := Permission{
		Name:        name,
		Description: description,
		Type:        permissionType,
		Actions:     actions,
		Prefix:      prefix,
		Buckets:     buckets,
		Policy:      policyJSON,
	}

	resp, err := conn.CreatePermission(&createPermissionInput)
	if err != nil {
		return fmt.Errorf("error creating permission: %w", err)
	}
	d.SetId(resp.ID)

	return resourcePermissionRead(d, meta)
}

func resourcePermissionRead(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(AccountAPI, meta.(Client)) {
		return fmt.Errorf("credentials for account api are missing")
	}

	conn := *meta.(Client).AccountAPIClient

	permissionId := d.Id()

	resource.RetryContext(context.Background(), time.Minute, func() *resource.RetryError {
		resp, err := conn.GetPermission(permissionId)

		if err == nil {
			d.Set("id", resp.Id)
			d.Set("name", resp.Name)
			d.Set("description", resp.Description)
			d.Set("type", resp.Type)

			if resp.Type != "policy" {
				d.Set("actions", resp.Actions)
			}

			if resp.Type == "bucket-names" {
				d.Set("buckets", resp.Buckets)
			}

			if resp.Type == "bucket-prefix" {
				d.Set("bucket_prefix", resp.Prefix)
			}

			d.Set("ready_state", resp.ReadyState)

			policy, err := unescape(resp.Policy)
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf("Error parsing policy: %s", err))
			}

			policyToSet, err := PolicyToSet(d.Get("policy").(string), policy)
			if err != nil {
				return resource.NonRetryableError(fmt.Errorf("Error setting policy: %s", err))
			}

			d.Set("policy", policyToSet)

			return nil
		}

		if err.Error() == InternalErr {
			return resource.RetryableError(fmt.Errorf("Error reading permission %s", permissionId))
		}

		if !d.IsNewResource() && err.Error() == PermissionNotFound {
			log.Printf("[WARN] Permission (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return resource.NonRetryableError(fmt.Errorf("Error reading permission: %s", err))
	})
	return nil
}

func resourcePermissionUpdate(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(AccountAPI, meta.(Client)) {
		return fmt.Errorf("credentials for account api are missing")
	}

	conn := *meta.(Client).AccountAPIClient

	permissionId := d.Id()

	name := NameWithSuffix(d.Get("name").(string), d.Get("name_prefix").(string))
	description := d.Get("description").(string)
	actions := d.Get("actions").(string)

	var policyJSON string
	var buckets []string
	prefix := d.Get("bucket_prefix").(string)

	var permissionType string

	if d.Get("all_buckets").(bool) {
		permissionType = "all-buckets"
	} else if prefix != "" {
		permissionType = "bucket-prefix"
	} else if v, ok := d.GetOk("buckets"); ok {
		permissionType = "bucket-names"
		var err error
		buckets, err = convertBucketsList(v.([]interface{}))
		if err != nil {
			return fmt.Errorf("error reading buckets list: %w", err)
		}
	} else if v, ok := d.GetOk("policy"); ok {
		permissionType = "policy"
		var err error
		policyJSON, err = NormalizeJsonString(v)
		if err != nil {
			return fmt.Errorf("policy (%s) is invalid JSON: %w", policyJSON, err)
		}
	} else {
		return fmt.Errorf("one of the following keys must be used: buckets/bucket_prefix/all_buckets/policy")
	}

	// update input for UpdatePermission
	updatePermissinInput := Permission{
		Name:        name,
		Description: description,
		Type:        permissionType,
		Actions:     actions,
		Prefix:      prefix,
		Buckets:     buckets,
		Policy:      policyJSON,
	}

	_, err := conn.UpdatePermission(permissionId, &updatePermissinInput)

	if !d.IsNewResource() && err.Error() == PermissionNotFound {
		log.Printf("[WARN] Permission (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error updating permission: %w", err)
	}

	return resourcePermissionRead(d, meta)
}

func resourcePermissionDelete(d *schema.ResourceData, meta interface{}) error {
	if CheckCredentials(AccountAPI, meta.(Client)) {
		return fmt.Errorf("credentials for account api are missing")
	}

	conn := *meta.(Client).AccountAPIClient

	_, err := conn.DeletePermission(d.Id())
	if err != nil {
		return fmt.Errorf("error deleting permission: %w", err)
	}

	return nil
}

// Takes a value containing JSON string and passes it through
// the JSON parser to normalize it, returns either a parsing
// error or normalized JSON string.
func NormalizeJsonString(jsonString interface{}) (string, error) {
	var j interface{}

	if jsonString == nil || jsonString.(string) == "" {
		return "", nil
	}

	s := jsonString.(string)

	err := json.Unmarshal([]byte(s), &j)
	if err != nil {
		return s, err
	}

	bytes, _ := json.Marshal(j)
	return string(bytes[:]), nil
}

// unescape unescapes a string
func unescape(s string) (string, error) {
	// Count %, check that they're well-formed.
	n := 0
	hasPlus := false
	for i := 0; i < len(s); {
		switch s[i] {
		case '%':
			n++
			if i+2 >= len(s) || !ishex(s[i+1]) || !ishex(s[i+2]) {
				s = s[i:]
				if len(s) > 3 {
					s = s[:3]
				}
				return "", EscapeError(s)
			}

			i += 3
		case '+':
			hasPlus = true
			i++
		default:
			i++
		}
	}

	if n == 0 && !hasPlus {
		return s, nil
	}

	var t strings.Builder
	t.Grow(len(s) - 2*n)
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '%':
			t.WriteByte(unhex(s[i+1])<<4 | unhex(s[i+2]))
			i += 2
		case '+':
			t.WriteByte(' ')
		default:
			t.WriteByte(s[i])
		}
	}
	return t.String(), nil
}

// PolicyToSet returns the existing policy if the new policy is equivalent.
// Otherwise, it returns the new policy. Either policy is normalized.
func PolicyToSet(exist, new string) (string, error) {
	policyToSet, err := SecondJSONUnlessEquivalent(exist, new)

	if err != nil {
		return "", fmt.Errorf("while checking equivalency of existing policy (%s) and new policy (%s), encountered: %w", exist, new, err)
	}

	policyToSet, err = NormalizeJsonString(policyToSet)

	if err != nil {
		return "", fmt.Errorf("policy (%s) is invalid JSON: %w", policyToSet, err)
	}

	return policyToSet, nil
}

func ishex(c byte) bool {
	switch {
	case '0' <= c && c <= '9':
		return true
	case 'a' <= c && c <= 'f':
		return true
	case 'A' <= c && c <= 'F':
		return true
	}
	return false
}

func (e EscapeError) Error() string {
	return "invalid URL escape " + strconv.Quote(string(e))
}

func unhex(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}

func SecondJSONUnlessEquivalent(old, new string) (string, error) {
	// valid empty JSON is "{}" not "" so handle special case to avoid
	// Error unmarshaling policy: unexpected end of JSON input
	if strings.TrimSpace(new) == "" {
		return "", nil
	}

	if strings.TrimSpace(new) == "{}" {
		return "{}", nil
	}

	if strings.TrimSpace(old) == "" || strings.TrimSpace(old) == "{}" {
		return new, nil
	}

	equivalent, err := awspolicy.PoliciesAreEquivalent(old, new)

	if err != nil {
		return "", err
	}

	if equivalent {
		return old, nil
	}

	return new, nil
}

// NameWithSuffix returns in order the name if non-empty, a prefix generated name if non-empty, or fully generated name prefixed with "terraform-".
func NameWithSuffix(name string, namePrefix string) string {
	if name != "" {
		return name
	}

	if namePrefix != "" {
		return resource.PrefixedUniqueId(namePrefix)
	}

	return resource.UniqueId()
}

func convertBucketsList(bucketsList []interface{}) ([]string, error) {
	buckets := []string{}
	for _, v := range bucketsList {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("error converting bucket: expected string, got %T", v)
		}
		buckets = append(buckets, str)
	}
	return buckets, nil
}
