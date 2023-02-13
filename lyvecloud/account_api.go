package lyvecloud

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// ErrorResponse holds the parsed response in case of error.
type ErrorResponse struct {
	Code    interface{} `json:"code,omitempty"`
	Message string      `json:"message"`
}

// AuthRequest specifies parameters for AuthAccountAPI.
type AuthRequest struct {
	AccountID string `json:"accountId"`
	AccessKey string `json:"accessKey"`
	Secret    string `json:"secret"`
}

// AuthData holds the response from the authentication request.
type AuthData struct {
	Token         string `json:"token"`
	ExpirationSec string `json:"expirationSec"`
}

// Permission specifies parameters for CreatePermission and UpdatePermission.
type Permission struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`    // should be auto computed in terraform. all-buckets/bucket-prefix/bucket-names/policy
	Actions     string   `json:"actions"` // all-operations/read-only/write-only
	Prefix      string   `json:"prefix"`
	Buckets     []string `json:"buckets"`
	Policy      string   `json:"policy"`
}

// ServiceAccount specifies parameters for CreateServiceAccount and UpdateServiceAccount.
type ServiceAccount struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// ServiceAccountResponse holds the parsed response from CreateServiceAccount.
type ServiceAccountResponse struct {
	ID        string `json:"id"`
	Accesskey string `json:"accessKey"`
	Secret    string `json:"secret"`
}

// PermissionResponse holds the parsed response from CreatePermission.
type PermissionResponse struct {
	ID string `json:"id"`
}

// GetPermissionResponse holds the parsed response from GetPermission.
type GetPermissionResponse struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	ReadyState  bool     `json:"readyState"`
	Actions     string   `json:"actions"`
	Prefix      string   `json:"prefix"`
	Buckets     []string `json:"buckets"`
	Policy      string   `json:"policy"`
}

// GetServiceAccountResponse holds the parsed response from GetServiceAccount.
type GetServiceAccountResponse struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Enabled     bool     `json:"enabled"`
	ReadyState  bool     `json:"readyState"`
	Permissions []string `json:"permissions"`
}

// The Dates structure holds the dates for which usage should be retrieved.
type Dates struct {
	fromMonth int
	fromYear  int
	toMonth   int
	toYear    int
}

// AuthAccountAPI returns access token.
func AuthAccountAPI(credentials *AuthRequest) (*AuthData, error) {
	payload, err := json.Marshal(credentials)
	if err != nil {
		return nil, err
	}

	resp, err := CreateAndSendRequest(http.MethodPost, TokenUrl, HeadersAuth(), bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var client *AuthData
	if err = json.Unmarshal(resBody, &client); err != nil {
		return nil, err
	}

	return client, nil
}

// CreatePermission creates permission.
func (c *AuthData) CreatePermission(permission *Permission) (*PermissionResponse, error) {
	payload, err := json.Marshal(permission)
	if err != nil {
		return nil, err
	}

	resp, err := CreateAndSendRequest(http.MethodPost, PermissionUrl, HeadersCreate(c), bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var permissionId *PermissionResponse
	if err = json.Unmarshal(resBody, &permissionId); err != nil {
		return nil, err
	}

	return permissionId, nil
}

// GetPermission retrieves given permission.
func (c *AuthData) GetPermission(permissionId string) (*GetPermissionResponse, error) {
	resp, err := CreateAndSendRequest(http.MethodGet, PermissionUrl+SlashSeparator+permissionId, HeadersGet(c), nil)
	if err != nil {
		return nil, err
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var getPermissionResp *GetPermissionResponse
	if err = json.Unmarshal(resBody, &getPermissionResp); err != nil {
		return nil, err
	}

	return getPermissionResp, nil
}

// DeletePermission deletes permission.
func (c *AuthData) DeletePermission(permissionId string) (int, error) {
	resp, err := CreateAndSendRequest(http.MethodDelete, PermissionUrl+SlashSeparator+permissionId, HeadersDelete(c), nil)
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}

// UpdatePermission updates permission.
func (c *AuthData) UpdatePermission(permissionId string, permission *Permission) (int, error) {
	payload, err := json.Marshal(permission)
	if err != nil {
		return 0, err
	}

	resp, err := CreateAndSendRequest(http.MethodPut, PermissionUrl+SlashSeparator+permissionId, HeadersCreate(c), bytes.NewBuffer(payload))
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}

// CreateServiceAccount creates service account.
func (c *AuthData) CreateServiceAccount(serviceAccount *ServiceAccount) (*ServiceAccountResponse, error) {
	payload, err := json.Marshal(serviceAccount)
	if err != nil {
		return nil, err
	}

	resp, err := CreateAndSendRequest(http.MethodPost, SAUrl, HeadersCreate(c), bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var serviceAccountData *ServiceAccountResponse
	if err := json.Unmarshal(resBody, &serviceAccountData); err != nil {
		return nil, err
	}

	return serviceAccountData, nil
}

func (c *AuthData) GetServiceAccount(serviceAccountId string) (*GetServiceAccountResponse, error) {
	resp, err := CreateAndSendRequest(http.MethodGet, SAUrl+SlashSeparator+serviceAccountId, HeadersGet(c), nil)
	if err != nil {
		return nil, err
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var getServiceAccountResp *GetServiceAccountResponse
	if err = json.Unmarshal(resBody, &getServiceAccountResp); err != nil {
		return nil, err
	}

	return getServiceAccountResp, nil
}

// UpdateServiceAccount updates given service account.
func (c *AuthData) UpdateServiceAccount(serviceAccountId string, serviceAccount *ServiceAccount) (int, error) {
	payload, err := json.Marshal(serviceAccount)
	if err != nil {
		return 0, err
	}

	resp, err := CreateAndSendRequest(http.MethodPut, SAUrl+SlashSeparator+serviceAccountId, HeadersCreate(c), bytes.NewBuffer(payload))
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}

// EnableServiceAccount enables service account.
func (c *AuthData) EnableServiceAccount(serviceAccountId string) (int, error) {
	resp, err := CreateAndSendRequest(http.MethodPut, SAUrl+SlashSeparator+serviceAccountId+SlashSeparator+Enabled, HeadersGet(c), nil)
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}

// DisableServiceAccoun disables service account.
func (c *AuthData) DisableServiceAccount(serviceAccountId string) (int, error) {
	resp, err := CreateAndSendRequest(http.MethodDelete, SAUrl+SlashSeparator+serviceAccountId+SlashSeparator+Enabled, HeadersGet(c), nil)
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}

// DeleteServiceAccount deletes service account.
func (c *AuthData) DeleteServiceAccount(serviceAccountId string) (int, error) {
	resp, err := CreateAndSendRequest(http.MethodDelete, SAUrl+SlashSeparator+serviceAccountId, HeadersDelete(c), nil)
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}

// GetUsageByDate returns the historical storage usage by month
func (c *AuthData) GetUsageByDate(dates Dates) (string, error) {
	// parse Dates struct to query string
	datesQuery := generateQueryString(dates)
	resp, err := CreateAndSendRequest(http.MethodGet, UsageMonthlyUrl+datesQuery, HeadersGet(c), nil)
	if err != nil {
		return "", err
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	jsonResponse := string(resBody)

	return jsonResponse, nil
}

// GetCurrentUsage returns the current month's storage usage in JSON string
func (c *AuthData) GetCurrentUsage() (string, error) {
	// parse Dates struct to string
	resp, err := CreateAndSendRequest(http.MethodGet, UsageCurrentUrl, HeadersGet(c), nil)
	if err != nil {
		return "", err
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	jsonResponse := string(resBody)

	return jsonResponse, nil
}

// CreateAndSendRequest creates http request and sends it.
func CreateAndSendRequest(method, url string, headers map[string][]string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header = headers

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		resp.Body.Close()

		// parse the JSON response into a Go struct
		var errResponse *ErrorResponse
		if err := json.Unmarshal([]byte(resBody), &errResponse); err != nil {
			return nil, err
		}

		if code, ok := errResponse.Code.(string); ok {
			return nil, errors.New(code)
		} else if code, ok := errResponse.Code.(int); ok {
			return nil, errors.New(strconv.Itoa(code))
		}
	}
	return resp, err
}

// generateQueryString takes a Dates struct and generates a query string based on its fields.
func generateQueryString(dates Dates) string {
	return fmt.Sprintf("?fromMonth=%d&fromYear=%d&toMonth=%d&toYear=%d",
		dates.fromMonth, dates.fromYear, dates.toMonth, dates.toYear)
}

// HeadersAuth returns headers for authorization.
func HeadersAuth() map[string][]string {
	return map[string][]string{
		ContentType: {Json},
		Accept:      {Json},
		UserAgent:   {TerraformProvider},
	}
}

// HeadersGet returns headers for disabling/enabling service account and retrieving permission/service account.
func HeadersGet(c *AuthData) map[string][]string {
	return map[string][]string{
		Accept:        {Json},
		Authorization: {Bearer + c.Token},
		UserAgent:     {TerraformProvider},
	}
}

// HeadersDelete returns headers for deleting permission/service account.
func HeadersDelete(c *AuthData) map[string][]string {
	return map[string][]string{
		Accept:        {Json},
		Authorization: {Bearer + c.Token},
		UserAgent:     {TerraformProvider},
	}
}

// HeadersDelete returns headers for creating permission/service account.
func HeadersCreate(c *AuthData) map[string][]string {
	return map[string][]string{
		Authorization: {Bearer + c.Token},
		ContentType:   {Json},
		Accept:        {Json},
		UserAgent:     {TerraformProvider},
	}
}
