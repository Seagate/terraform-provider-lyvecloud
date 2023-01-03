package lyvecloud

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// AuthV2 specifies parameters for AuthAccountAPIV2.
type AuthV2 struct {
	AccountID string `json:"accountId"`
	AccessKey string `json:"accessKey"`
	Secret    string `json:"secret"`
}

// AuthDataV2 holds the response from the authentication request.
type AuthDataV2 struct {
	Token         string `json:"token"`
	ExpirationSec string `json:"expirationSec"`
}

// PermissionV2 specifies parameters for CreatePermissionV2 and UpdatePermissionV2.
type PermissionV2 struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`    // should be auto computed in terraform. all-buckets/bucket-prefix/bucket-names/policy
	Actions     string   `json:"actions"` // all-operations/read-only/write-only
	Prefix      string   `json:"prefix"`
	Buckets     []string `json:"buckets"`
}

// ServiceAccountResponseV2 holds the parsed response from CreateServiceAccount.
type ServiceAccountResponseV2 struct {
	ID        string
	Accesskey string
	Secret    string
}

// GetPermissionResponseV2 holds the parsed response from GetPermissionV2.
type GetPermissionResponseV2 struct {
	Id          string
	Name        string
	Description string
	Type        string
	ReadyState  bool
	Actions     string
	Prefix      string
	Buckets     []string
}

// GetServiceAccountResponseV2 holds the parsed response from GetServiceAccountV2.
type GetServiceAccountResponseV2 struct {
	Id          string
	Name        string
	Description string
	Enabled     bool
	ReadyState  bool
	Permissions []string
}

// AuthAccountAPIV2 returns access token.
func AuthAccountAPIV2(credentials *AuthV2) (*AuthDataV2, error) {
	var client *AuthDataV2

	payload, err := json.Marshal(credentials)
	if err != nil {
		return nil, err
	}

	resp, err := CreateAndSendRequest(http.MethodPost, TokenUrlV2STG, HeadersAuthV2(), bytes.NewBuffer(payload))

	if err != nil {
		return client, err
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return client, err
	}

	if err = json.Unmarshal(resBody, &client); err != nil {
		return client, err
	}

	if err != nil {
		return client, err
	}

	return client, nil
}

// CreatePermissionV2 creates permission.
func (c *AuthDataV2) CreatePermissionV2(permission *PermissionV2) (*PermissionResponse, error) {
	var pid *PermissionResponse

	payload, err := json.Marshal(permission)
	if err != nil {
		return pid, err
	}

	resp, err := CreateAndSendRequest(http.MethodPost, PermissionUrlV2STG, HeadersCreateV2(c), bytes.NewBuffer(payload))
	if err != nil {
		return pid, err
	}
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return pid, err
	}

	if err = json.Unmarshal(resBody, &pid); err != nil {
		return pid, err
	}

	if err != nil {
		return pid, err
	}

	return pid, nil
}

// GetPermissionV2 retrieves given permission.
func (c *AuthDataV2) GetPermissionV2(permissionId string) (*GetPermissionResponseV2, error) {
	var getPermissionResp *GetPermissionResponseV2

	resp, err := CreateAndSendRequest(http.MethodGet, PermissionUrlV2STG+SlashSeparator+permissionId, HeadersGetV2(c), nil)
	if err != nil {
		return getPermissionResp, err
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return getPermissionResp, err
	}

	if err = json.Unmarshal(resBody, &getPermissionResp); err != nil {
		return getPermissionResp, err
	}

	return getPermissionResp, nil
}

// DeletePermissionV2 deletes permission.
func (c *AuthDataV2) DeletePermissionV2(permissionId string) (*http.Response, error) {
	return CreateAndSendRequest(http.MethodDelete, PermissionUrlV2STG+SlashSeparator+permissionId, HeadersDeleteV2(c), nil)
}

// UpdatePermissionV2 updates permission.
func (c *AuthDataV2) UpdatePermissionV2(permissionId string, permission *PermissionV2) (*http.Response, error) {
	payload, err := json.Marshal(permission)
	if err != nil {
		return nil, err
	}

	return CreateAndSendRequest(http.MethodPut, PermissionUrlV2STG+SlashSeparator+permissionId, HeadersCreateV2(c), bytes.NewBuffer(payload))
}

// CreateServiceAccountV2 creates service account.
func (c *AuthDataV2) CreateServiceAccountV2(serviceAccount *ServiceAccount) (*ServiceAccountResponseV2, error) {
	var sad *ServiceAccountResponseV2
	payload, err := json.Marshal(serviceAccount)
	if err != nil {
		return sad, err
	}

	resp, err := CreateAndSendRequest(http.MethodPost, SAUrlV2STG, HeadersCreateV2(c), bytes.NewBuffer(payload))
	if err != nil {
		return sad, err
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return sad, err
	}

	if err = json.Unmarshal(resBody, &sad); err != nil {
		return sad, err
	}

	if err != nil {
		return sad, err
	}

	return sad, nil
}

// GetServiceAccountV2 retrieves given service account.
func (c *AuthDataV2) GetServiceAccountV2(serviceAccountId string) (*GetServiceAccountResponseV2, error) {
	var getServiceAccountResp *GetServiceAccountResponseV2

	resp, err := CreateAndSendRequest(http.MethodGet, SAUrlV2STG+SlashSeparator+serviceAccountId, HeadersGetV2(c), nil)
	if err != nil {
		return getServiceAccountResp, err
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return getServiceAccountResp, err
	}

	if err = json.Unmarshal(resBody, &getServiceAccountResp); err != nil {
		return getServiceAccountResp, err
	}

	return getServiceAccountResp, nil
}

// UpdateServiceAccountV2 updates given service account.
func (c *AuthDataV2) UpdateServiceAccountV2(serviceAccountId string, serviceAccount *ServiceAccount) (*ServiceAccountResponse, error) {
	var sad *ServiceAccountResponse

	payload, err := json.Marshal(serviceAccount)
	if err != nil {
		return sad, err
	}

	resp, err := CreateAndSendRequest(http.MethodPut, SAUrlV2STG+SlashSeparator+serviceAccountId, HeadersCreateV2(c), bytes.NewBuffer(payload))
	if err != nil {
		return sad, err
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return sad, err
	}

	if err = json.Unmarshal(resBody, &sad); err != nil {
		return sad, err
	}

	if err != nil {
		return sad, err
	}

	return sad, nil
}

// EnableServiceAccountV2 enables service account.
func (c *AuthDataV2) EnableServiceAccountV2(serviceAccountId string) (*http.Response, error) {
	return CreateAndSendRequest(http.MethodPut, SAUrlV2STG+SlashSeparator+serviceAccountId+SlashSeparator+Enabled, HeadersGetV2(c), nil)
}

// DisableServiceAccountV2 disables service account.
func (c *AuthDataV2) DisableServiceAccountV2(serviceAccountId string) (*http.Response, error) {
	return CreateAndSendRequest(http.MethodDelete, SAUrlV2STG+SlashSeparator+serviceAccountId+SlashSeparator+Enabled, HeadersGetV2(c), nil)
}

// DeleteServiceAccountV2 deletes service account.
func (c *AuthDataV2) DeleteServiceAccountV2(serviceAccountId string) (*http.Response, error) {
	return CreateAndSendRequest(http.MethodDelete, SAUrlV2STG+SlashSeparator+serviceAccountId, HeadersDeleteV2(c), nil)
}

// HeadersAuthV2 returns headers for authorization.
func HeadersAuthV2() map[string][]string {
	return map[string][]string{
		ContentType: {
			Json,
		},
		Accept: {
			Json,
		},
	}
}

// HeadersGetV2 returns headers for disabling/enabling service account and retrieving permission/service account.
func HeadersGetV2(c *AuthDataV2) map[string][]string {
	return map[string][]string{
		Accept:        {Json},
		Authorization: {Bearer + c.Token},
	}
}

// HeadersDeleteV2 returns headers for deleting permission/service account.
func HeadersDeleteV2(c *AuthDataV2) map[string][]string {
	return map[string][]string{
		Accept:        {Json},
		Authorization: {Bearer + c.Token},
	}
}

// HeadersDeleteV2 returns headers for creating permission/service account.
func HeadersCreateV2(c *AuthDataV2) map[string][]string {
	return map[string][]string{
		Authorization: {
			Bearer + c.Token,
		},
		ContentType: {
			Json,
		},
		Accept: {
			Json,
		},
	}
}
