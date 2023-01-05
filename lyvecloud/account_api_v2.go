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
	ID        string `json:"id"`
	Accesskey string `json:"accessKey"`
	Secret    string `json:"secret"`
}

// GetPermissionResponseV2 holds the parsed response from GetPermissionV2.
type GetPermissionResponseV2 struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	ReadyState  bool     `json:"readyState"`
	Actions     string   `json:"actions"`
	Prefix      string   `json:"prefix"`
	Buckets     []string `json:"buckets"`
}

// GetServiceAccountResponseV2 holds the parsed response from GetServiceAccountV2.
type GetServiceAccountResponseV2 struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Enabled     bool     `json:"enabled"`
	ReadyState  bool     `json:"readyState"`
	Permissions []string `json:"permissions"`
}

// AuthAccountAPIV2 returns access token.
func AuthAccountAPIV2(credentials *AuthV2) (*AuthDataV2, error) {
	payload, err := json.Marshal(credentials)
	if err != nil {
		return nil, err
	}

	resp, err := CreateAndSendRequest(http.MethodPost, TokenUrlV2STG, HeadersAuthV2(), bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var client *AuthDataV2
	if err = json.Unmarshal(resBody, client); err != nil {
		return nil, err
	}

	return client, nil
}

// CreatePermissionV2 creates permission.
func (c *AuthDataV2) CreatePermissionV2(permission *PermissionV2) (*PermissionResponse, error) {
	payload, err := json.Marshal(permission)
	if err != nil {
		return nil, err
	}

	resp, err := CreateAndSendRequest(http.MethodPost, PermissionUrlV2STG, HeadersCreateV2(c), bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var permissionId *PermissionResponse
	if err = json.Unmarshal(resBody, permissionId); err != nil {
		return nil, err
	}

	return permissionId, nil
}

// GetPermissionV2 retrieves given permission.
func (c *AuthDataV2) GetPermissionV2(permissionId string) (*GetPermissionResponseV2, error) {
	resp, err := CreateAndSendRequest(http.MethodGet, PermissionUrlV2STG+SlashSeparator+permissionId, HeadersGetV2(c), nil)
	if err != nil {
		return nil, err
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var getPermissionResp *GetPermissionResponseV2
	if err = json.Unmarshal(resBody, getPermissionResp); err != nil {
		return nil, err
	}

	return getPermissionResp, nil
}

// DeletePermissionV2 deletes permission.
func (c *AuthDataV2) DeletePermissionV2(permissionId string) (int, error) {
	resp, err := CreateAndSendRequest(http.MethodDelete, PermissionUrlV2STG+SlashSeparator+permissionId, HeadersDeleteV2(c), nil)
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}

// UpdatePermissionV2 updates permission.
func (c *AuthDataV2) UpdatePermissionV2(permissionId string, permission *PermissionV2) (int, error) {
	payload, err := json.Marshal(permission)
	if err != nil {
		return 0, err
	}

	resp, err := CreateAndSendRequest(http.MethodPut, PermissionUrlV2STG+SlashSeparator+permissionId, HeadersCreateV2(c), bytes.NewBuffer(payload))
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}

// CreateServiceAccountV2 creates service account.
func (c *AuthDataV2) CreateServiceAccountV2(serviceAccount *ServiceAccount) (*ServiceAccountResponseV2, error) {
	payload, err := json.Marshal(serviceAccount)
	if err != nil {
		return nil, err
	}

	resp, err := CreateAndSendRequest(http.MethodPost, SAUrlV2STG, HeadersCreateV2(c), bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var serviceAccountData *ServiceAccountResponseV2
	if err := json.Unmarshal(resBody, serviceAccountData); err != nil {
		return nil, err
	}

	return serviceAccountData, nil
}

func (c *AuthDataV2) GetServiceAccountV2(serviceAccountId string) (*GetServiceAccountResponseV2, error) {
	resp, err := CreateAndSendRequest(http.MethodGet, SAUrlV2STG+SlashSeparator+serviceAccountId, HeadersGetV2(c), nil)
	if err != nil {
		return nil, err
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var getServiceAccountResp *GetServiceAccountResponseV2
	if err = json.Unmarshal(resBody, getServiceAccountResp); err != nil {
		return nil, err
	}

	return getServiceAccountResp, nil
}

// UpdateServiceAccountV2 updates given service account.
func (c *AuthDataV2) UpdateServiceAccountV2(serviceAccountId string, serviceAccount *ServiceAccount) (*ServiceAccountResponse, error) {
	var serviceAccountData *ServiceAccountResponse

	payload, err := json.Marshal(serviceAccount)
	if err != nil {
		return nil, err
	}

	resp, err := CreateAndSendRequest(http.MethodPut, SAUrlV2STG+SlashSeparator+serviceAccountId, HeadersCreateV2(c), bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err = json.Unmarshal(resBody, serviceAccountData); err != nil {
		return nil, err
	}

	return serviceAccountData, nil
}

// EnableServiceAccountV2 enables service account.
func (c *AuthDataV2) EnableServiceAccountV2(serviceAccountId string) (int, error) {
	resp, err := CreateAndSendRequest(http.MethodPut, SAUrlV2STG+SlashSeparator+serviceAccountId+SlashSeparator+Enabled, HeadersGetV2(c), nil)
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}

// DisableServiceAccountV2 disables service account.
func (c *AuthDataV2) DisableServiceAccountV2(serviceAccountId string) (int, error) {
	resp, err := CreateAndSendRequest(http.MethodDelete, SAUrlV2STG+SlashSeparator+serviceAccountId+SlashSeparator+Enabled, HeadersGetV2(c), nil)
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}

// DeleteServiceAccountV2 deletes service account.
func (c *AuthDataV2) DeleteServiceAccountV2(serviceAccountId string) (int, error) {
	resp, err := CreateAndSendRequest(http.MethodDelete, SAUrlV2STG+SlashSeparator+serviceAccountId, HeadersDeleteV2(c), nil)
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}

// HeadersAuthV2 returns headers for authorization.
func HeadersAuthV2() map[string][]string {
	return map[string][]string{
		ContentType: {Json},
		Accept:      {Json},
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
		Authorization: {Bearer + c.Token},
		ContentType:   {Json},
		Accept:        {Json},
	}
}
