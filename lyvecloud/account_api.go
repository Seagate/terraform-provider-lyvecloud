package lyvecloud

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// Auth specifies parameters for AuthAccountAPI.
type Auth struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Audience     string `json:"audience"`
	GrantType    string `json:"grant_type"`
}

// AuthData holds the response from the authentication request.
type AuthData struct {
	Access_token string
	Expires_in   int
	Token_type   string
}

// Permission specifies parameters for CreatePermission.
type Permission struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Actions     string   `json:"actions"` // all-operations/read/write
	Buckets     []string `json:"buckets"`
}

// PermissionResponse holds the parsed response from CreatePermission.
type PermissionResponse struct {
	ID string
}

// ServiceAccount specifies parameters for CreateServiceAccount.
type ServiceAccount struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// ServiceAccountResponse holds the parsed response from CreateServiceAccount.
type ServiceAccountResponse struct {
	ID            string
	Access_key    string
	Access_Secret string
}

// ErrorResponse holds the parsed response in case of error.
type ErrorResponse struct {
	Error   string      `json:"error"`
	Code    interface{} `json:"code,omitempty"`
	Message string      `json:"message"`
}

// AuthAccountAPI returns access token.
func AuthAccountAPI(credentials *Auth) (*AuthData, error) {
	var client *AuthData

	credentials.Audience = AudienceUrl
	credentials.GrantType = ClientCredentials

	payload, err := json.Marshal(credentials)
	if err != nil {
		return nil, err
	}

	resp, err := CreateAndSendRequest(http.MethodPost, TokenUrl, HeadersAuth(), bytes.NewBuffer(payload))
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

// CreatePermission creates permission.
func (c *AuthData) CreatePermission(permission *Permission) (*PermissionResponse, error) {
	var pid *PermissionResponse

	payload, err := json.Marshal(permission)
	if err != nil {
		return pid, err
	}

	resp, err := CreateAndSendRequest(http.MethodPut, PermissionUrl, HeadersCreate(c), bytes.NewBuffer(payload))
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

// DeletePermission deletes permission.
func (c *AuthData) DeletePermission(permissionId string) (*http.Response, error) {
	return CreateAndSendRequest(http.MethodDelete, PermissionUrl+SlashSeparator+permissionId, HeadersDelete(c), nil)
}

// CreateServiceAccount creates service account.
func (c *AuthData) CreateServiceAccount(serviceAccount *ServiceAccount) (*ServiceAccountResponse, error) {
	var sad *ServiceAccountResponse
	payload, err := json.Marshal(serviceAccount)
	if err != nil {
		return sad, err
	}

	resp, err := CreateAndSendRequest(http.MethodPut, SAUrl, HeadersCreate(c), bytes.NewBuffer(payload))
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

// DeleteServiceAccount deletes service account.
func (c *AuthData) DeleteServiceAccount(permissionId string) (*http.Response, error) {
	return CreateAndSendRequest(http.MethodDelete, SAUrl+SlashSeparator+permissionId, HeadersDelete(c), nil)
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

	if resp.StatusCode != 200 {
		resBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		// parse the JSON response into a Go struct
		var response ErrorResponse
		if err := json.Unmarshal([]byte(resBody), &response); err != nil {
			return nil, err
		}

		// check the status code and return either the "error" field or the "message" field
		if response.Error != "" {
			return nil, errors.New(response.Error)
		} else {
			return nil, errors.New(response.Message)
		}
	}

	return resp, err
}

// HeadersCreate returns headers for creating permission/service account.
func HeadersCreate(c *AuthData) map[string][]string {
	return map[string][]string{
		Authorization: {Bearer + c.Access_token},
	}
}

// HeadersDelete returns headers for deleting permission/service account.
func HeadersDelete(c *AuthData) map[string][]string {
	return map[string][]string{
		ContentType:   {Json},
		Authorization: {Bearer + c.Access_token},
	}
}

// HeadersAuth returns headers for account api v2 authorization.
func HeadersAuth() map[string][]string {
	return map[string][]string{
		ContentType: {Json},
	}
}
