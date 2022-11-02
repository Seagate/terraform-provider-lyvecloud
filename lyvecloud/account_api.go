package lyvecloud

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ErrorResponse holds the parsed response from in case if account api returns error.
type ErrorResponse struct {
	Error string
}

// PermissionResponse holds the parsed response from CreatePermission.
type PermissionResponse struct {
	ID string
}

// ServiceAccountResponse holds the parsed response from CreateServiceAccount.
type ServiceAccountResponse struct {
	ID            string
	Access_key    string
	Access_Secret string
}

// Permission specifies parameters for CreatePermission.
type Permission struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Actions     string   `json:"actions"`
	Buckets     []string `json:"buckets"`
}

// ServiceAccount specifies parameters for CreateServiceAccount.
type ServiceAccount struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// AuthAccountAPI returns access token.
func AuthAccountAPI(clientId, clientSecret string) (*AuthData, error) {
	var client *AuthData
	r := fmt.Sprintf(ClientReq, clientId, clientSecret)
	resp, err := CreateAndSendRequest(Post, TokenUrl, map[string]string{ContentType: Json}, strings.NewReader(r))
	if err != nil {
		return client, err
	}

	if resp.StatusCode == ErrCodeUnauthorized {
		return client, errors.New(UnauthorizedMessage)
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
func (c *AuthData) CreatePermission(name, desc, actions string, buckets []string) (*PermissionResponse, error) {
	var pid *PermissionResponse
	data := SetPermission(name, desc, actions, buckets)
	buf, err := json.Marshal(data)
	if err != nil {
		return pid, err
	}

	resp, err := CreateAndSendRequest(Put, PermissionUrl, map[string]string{Authorization: Bearer + c.Access_token, ContentType: Json}, bytes.NewReader(buf))
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
	return CreateAndSendRequest(Delete, PermissionUrl+SlashSeparator+permissionId, map[string]string{Authorization: Bearer + c.Access_token}, nil)
}

// CreateServiceAccount creates service account.
func (c *AuthData) CreateServiceAccount(name, desc string, permissions []string) (*ServiceAccountResponse, error) {
	var sad *ServiceAccountResponse
	data := SetSA(name, desc, permissions)
	buf, err := json.Marshal(data)
	if err != nil {
		return sad, err
	}

	resp, err := CreateAndSendRequest(Put, SAUrl, map[string]string{Authorization: Bearer + c.Access_token, ContentType: Json}, bytes.NewReader(buf))
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
	return CreateAndSendRequest(Delete, SAUrl+SlashSeparator+permissionId, map[string]string{Authorization: Bearer + c.Access_token}, nil)
}

// CreateAndSendRequest creates http request and sends it.
func CreateAndSendRequest(method, url string, m map[string]string, body io.Reader) (*http.Response, error) {

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, errors.New(ErrBadRequest)
	}

	for key, val := range m {
		req.Header.Set(key, val)
	}

	resp, err := http.DefaultClient.Do(req)

	if resp.StatusCode != 200 {
		var errMesg *ErrorResponse
		resBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		if err = json.Unmarshal(resBody, &errMesg); err != nil {
			return nil, err
		}

		if err != nil {
			return nil, err
		}

		return nil, errors.New(errMesg.Error)
	}

	return resp, err
}

// SetPermission initializes a struct for the http request that creates permission.
func SetPermission(name, desc, actions string, buckets []string) *Permission {
	return &Permission{
		Name:        name,
		Description: desc,
		Actions:     actions,
		Buckets:     buckets,
	}
}

// SetSA initializes a struct for the http request that creates service account.
func SetSA(name, desc string, permissions []string) *ServiceAccount {
	return &ServiceAccount{
		Name:        name,
		Description: desc,
		Permissions: permissions,
	}
}
