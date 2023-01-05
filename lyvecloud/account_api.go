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
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"Token_type"`
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
	ID string `json:"id"`
}

// ServiceAccount specifies parameters for CreateServiceAccount.
type ServiceAccount struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// ServiceAccountResponse holds the parsed response from CreateServiceAccount.
type ServiceAccountResponse struct {
	ID           string `json:"Id"`
	Accesskey    string `json:"accessKey"`
	AccessSecret string `json:"access_secret"`
}

// ErrorResponse holds the parsed response in case of error.
type ErrorResponse struct {
	Error   string      `json:"error"`
	Code    interface{} `json:"code,omitempty"`
	Message string      `json:"message"`
}

// AuthAccountAPI returns access token.
func AuthAccountAPI(credentials *Auth) (*AuthData, error) {
	credentials.Audience = AudienceUrl
	credentials.GrantType = ClientCredentials

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
	defer resp.Body.Close()

	var client *AuthData
	if err = json.Unmarshal(resBody, client); err != nil {
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

	resp, err := CreateAndSendRequest(http.MethodPut, PermissionUrl, HeadersCreate(c), bytes.NewBuffer(payload))
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

// DeletePermission deletes permission.
func (c *AuthData) DeletePermission(permissionId string) (int, error) {
	resp, err := CreateAndSendRequest(http.MethodDelete, PermissionUrl+SlashSeparator+permissionId, HeadersDelete(c), nil)
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

	resp, err := CreateAndSendRequest(http.MethodPut, SAUrl, HeadersCreate(c), bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var serviceAccountData *ServiceAccountResponse
	if err = json.Unmarshal(resBody, serviceAccountData); err != nil {
		return nil, err
	}

	return serviceAccountData, nil
}

func (c *AuthData) DeleteServiceAccount(serviceAccountId string) (statusCode int, err error) {
	resp, err := CreateAndSendRequest(http.MethodDelete, SAUrl+SlashSeparator+serviceAccountId, HeadersDelete(c), nil)
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
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
		if err := json.Unmarshal([]byte(resBody), errResponse); err != nil {
			return nil, err
		}

		// check the status code and return either the "error" field or the "message" field
		if errResponse.Error != "" {
			return nil, errors.New(errResponse.Error)
		} else {
			return nil, errors.New(errResponse.Message)
		}
	}

	return resp, err
}

// HeadersCreate returns headers for creating permission/service account.
func HeadersCreate(c *AuthData) map[string][]string {
	return map[string][]string{
		Authorization: {Bearer + c.AccessToken},
	}
}

// HeadersDelete returns headers for deleting permission/service account.
func HeadersDelete(c *AuthData) map[string][]string {
	return map[string][]string{
		ContentType:   {Json},
		Authorization: {Bearer + c.AccessToken},
	}
}

// HeadersAuth returns headers for account api v2 authorization.
func HeadersAuth() map[string][]string {
	return map[string][]string{
		ContentType: {Json},
	}
}
