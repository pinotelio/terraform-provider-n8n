package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is the n8n API client
type Client struct {
	HTTPClient *http.Client
	BaseURL    string
	APIKey     string
}

// NewClient creates a new n8n API client
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: strings.TrimSuffix(baseURL, "/"),
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	url := fmt.Sprintf("%s%s", c.BaseURL, path)
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-N8N-API-KEY", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log the error but don't override the main error
			_ = closeErr
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Workflow represents an n8n workflow
type Workflow struct {
	Connections map[string]interface{} `json:"connections"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
	ID          string                 `json:"id,omitempty"`
	Name        string                 `json:"name"`
	CreatedAt   string                 `json:"createdAt,omitempty"`
	UpdatedAt   string                 `json:"updatedAt,omitempty"`
	Nodes       []interface{}          `json:"nodes"`
	Tags        []map[string]string    `json:"tags,omitempty"`
	Active      bool                   `json:"active"`
}

// WorkflowListResponse represents the response from listing workflows
type WorkflowListResponse struct {
	Data []Workflow `json:"data"`
}

// CreateWorkflow creates a new workflow
func (c *Client) CreateWorkflow(workflow *Workflow) (*Workflow, error) {
	// Store the desired tags (read-only on creation)
	// Note: active field is now managed by n8n_workflow_activation resource
	desiredTags := workflow.Tags

	// Create workflow without tags field (it's read-only on creation)
	createPayload := map[string]interface{}{
		"name":        workflow.Name,
		"nodes":       workflow.Nodes,
		"connections": workflow.Connections,
	}

	if workflow.Settings != nil {
		createPayload["settings"] = workflow.Settings
	}

	respBody, err := c.doRequest("POST", "/api/v1/workflows", createPayload)
	if err != nil {
		return nil, err
	}

	var result Workflow
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// If tags are specified, update them after creation
	// Only update if tags have actual content (not just empty array from n8n export)
	if len(desiredTags) > 0 {
		// Check if tags have valid IDs
		hasValidTags := false
		for _, tag := range desiredTags {
			if id, ok := tag["id"]; ok && id != "" {
				hasValidTags = true
				break
			}
		}

		if hasValidTags {
			if err := c.UpdateWorkflowTags(result.ID, desiredTags); err != nil {
				// If tags update fails, delete the workflow to clean up
				deleteErr := c.DeleteWorkflow(result.ID)
				if deleteErr != nil {
					return nil, fmt.Errorf("failed to update workflow tags: %w (also failed to clean up workflow: %v) - hint: tags must exist in n8n before assigning them to workflows", err, deleteErr)
				}
				return nil, fmt.Errorf("failed to update workflow tags, workflow rolled back: %w (hint: tags must exist in n8n before assigning them to workflows)", err)
			}
			result.Tags = desiredTags
		}
	}

	return &result, nil
}

// GetWorkflow retrieves a workflow by ID
func (c *Client) GetWorkflow(id string) (*Workflow, error) {
	respBody, err := c.doRequest("GET", fmt.Sprintf("/api/v1/workflows/%s", id), nil)
	if err != nil {
		return nil, err
	}

	var result Workflow
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// UpdateWorkflow updates an existing workflow
func (c *Client) UpdateWorkflow(id string, workflow *Workflow) (*Workflow, error) {
	// Store the desired tags (read-only)
	// Note: active field is now managed by n8n_workflow_activation resource
	desiredTags := workflow.Tags

	// Update workflow without tags field (it's read-only)
	updatePayload := map[string]interface{}{
		"name":        workflow.Name,
		"nodes":       workflow.Nodes,
		"connections": workflow.Connections,
	}

	if workflow.Settings != nil {
		updatePayload["settings"] = workflow.Settings
	}

	respBody, err := c.doRequest("PUT", fmt.Sprintf("/api/v1/workflows/%s", id), updatePayload)
	if err != nil {
		return nil, err
	}

	var result Workflow
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Update tags if they changed
	if len(desiredTags) > 0 {
		// Check if tags have valid IDs
		hasValidTags := false
		for _, tag := range desiredTags {
			if id, ok := tag["id"]; ok && id != "" {
				hasValidTags = true
				break
			}
		}

		if hasValidTags {
			if err := c.UpdateWorkflowTags(id, desiredTags); err != nil {
				return nil, fmt.Errorf("failed to update workflow tags: %w (hint: tags must exist in n8n before assigning them to workflows)", err)
			}
			result.Tags = desiredTags
		}
	}

	return &result, nil
}

// DeleteWorkflow deletes a workflow
func (c *Client) DeleteWorkflow(id string) error {
	_, err := c.doRequest("DELETE", fmt.Sprintf("/api/v1/workflows/%s", id), nil)
	return err
}

// ActivateWorkflow activates a workflow
func (c *Client) ActivateWorkflow(id string) (*Workflow, error) {
	respBody, err := c.doRequest("POST", fmt.Sprintf("/api/v1/workflows/%s/activate", id), nil)
	if err != nil {
		return nil, err
	}

	var result Workflow
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// DeactivateWorkflow deactivates a workflow
func (c *Client) DeactivateWorkflow(id string) (*Workflow, error) {
	respBody, err := c.doRequest("POST", fmt.Sprintf("/api/v1/workflows/%s/deactivate", id), nil)
	if err != nil {
		return nil, err
	}

	var result Workflow
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// UpdateWorkflowTags updates the tags of a workflow
func (c *Client) UpdateWorkflowTags(id string, tags []map[string]string) error {
	// Convert tags to the format expected by the API
	tagPayload := make([]map[string]string, len(tags))
	for i, tag := range tags {
		tagPayload[i] = map[string]string{
			"id": tag["id"],
		}
	}

	_, err := c.doRequest("PUT", fmt.Sprintf("/api/v1/workflows/%s/tags", id), tagPayload)
	return err
}

// ListWorkflows lists all workflows
func (c *Client) ListWorkflows() ([]Workflow, error) {
	respBody, err := c.doRequest("GET", "/api/v1/workflows", nil)
	if err != nil {
		return nil, err
	}

	var result WorkflowListResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.Data, nil
}

// Credential represents an n8n credential
type Credential struct {
	Data map[string]interface{} `json:"data,omitempty"`
	ID   string                 `json:"id,omitempty"`
	Name string                 `json:"name"`
	Type string                 `json:"type"`
}

// CredentialListResponse represents the response from listing credentials
type CredentialListResponse struct {
	Data []Credential `json:"data"`
}

// CreateCredential creates a new credential
func (c *Client) CreateCredential(credential *Credential) (*Credential, error) {
	respBody, err := c.doRequest("POST", "/api/v1/credentials", credential)
	if err != nil {
		return nil, err
	}

	var result Credential
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// GetCredential retrieves a credential by ID
func (c *Client) GetCredential(id string) (*Credential, error) {
	respBody, err := c.doRequest("GET", fmt.Sprintf("/api/v1/credentials/%s", id), nil)
	if err != nil {
		return nil, err
	}

	var result Credential
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// UpdateCredential updates an existing credential
func (c *Client) UpdateCredential(id string, credential *Credential) (*Credential, error) {
	respBody, err := c.doRequest("PATCH", fmt.Sprintf("/api/v1/credentials/%s", id), credential)
	if err != nil {
		return nil, err
	}

	var result Credential
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// DeleteCredential deletes a credential
func (c *Client) DeleteCredential(id string) error {
	_, err := c.doRequest("DELETE", fmt.Sprintf("/api/v1/credentials/%s", id), nil)
	return err
}

// ListCredentials lists all credentials
func (c *Client) ListCredentials() ([]Credential, error) {
	respBody, err := c.doRequest("GET", "/api/v1/credentials", nil)
	if err != nil {
		return nil, err
	}

	var result CredentialListResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.Data, nil
}

// User represents an n8n user
type User struct {
	ID              string `json:"id,omitempty"`
	Email           string `json:"email"`
	Role            string `json:"role,omitempty"`
	GlobalRole      string `json:"globalRole,omitempty"` // Some n8n versions use globalRole instead of role
	CreatedAt       string `json:"createdAt,omitempty"`
	UpdatedAt       string `json:"updatedAt,omitempty"`
	InviteAcceptURL string `json:"inviteAcceptUrl,omitempty"` // Only populated on user creation
	IsOwner         bool   `json:"isOwner,omitempty"`
	IsPending       bool   `json:"isPending,omitempty"`
}

// GetRole returns the role, preferring GlobalRole if Role is empty
func (u *User) GetRole() string {
	if u.Role != "" {
		return u.Role
	}
	return u.GlobalRole
}

// SetRole sets both Role and GlobalRole to ensure compatibility
func (u *User) SetRole(role string) {
	u.Role = role
	u.GlobalRole = role
}

// UserListResponse represents the response from listing users
type UserListResponse struct {
	Data []User `json:"data"`
}

// CreateUserResponse represents the response from creating users
type CreateUserResponse struct {
	Error string `json:"error"`
	User  struct {
		ID              string `json:"id"`
		Email           string `json:"email"`
		InviteAcceptURL string `json:"inviteAcceptUrl,omitempty"`
		Role            string `json:"role,omitempty"`
		EmailSent       bool   `json:"emailSent,omitempty"`
	} `json:"user"`
}

// CreateUser creates a new user
func (c *Client) CreateUser(user *User) (*User, error) {
	// n8n API expects an array of users for bulk creation
	// The request should only include email and role
	type CreateUserRequest struct {
		Email string `json:"email"`
		Role  string `json:"role,omitempty"`
	}

	request := CreateUserRequest{
		Email: user.Email,
		Role:  user.Role,
	}

	users := []CreateUserRequest{request}
	respBody, err := c.doRequest("POST", "/api/v1/users", users)
	if err != nil {
		return nil, err
	}

	// The response is an array of objects with "user" and "error" fields
	var results []CreateUserResponse
	if err := json.Unmarshal(respBody, &results); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no user returned from API")
	}

	if results[0].Error != "" {
		return nil, fmt.Errorf("API error: %s", results[0].Error)
	}

	// Preserve the inviteAcceptUrl from the creation response
	inviteAcceptURL := results[0].User.InviteAcceptURL

	// Fetch the full user details to get all fields including role, timestamps, etc.
	// The create response doesn't include all fields we need
	createdUser, err := c.GetUser(results[0].User.ID)
	if err != nil {
		return nil, err
	}

	// If the API doesn't return the role in GetUser response, preserve the role from the request
	// This handles cases where n8n API doesn't return role/globalRole in the GET response
	if createdUser.GetRole() == "" && user.Role != "" {
		createdUser.SetRole(user.Role)
	}

	// Set the inviteAcceptUrl from the creation response (not available in GET response)
	createdUser.InviteAcceptURL = inviteAcceptURL

	return createdUser, nil
}

// GetUser retrieves a user by ID
func (c *Client) GetUser(id string) (*User, error) {
	respBody, err := c.doRequest("GET", fmt.Sprintf("/api/v1/users/%s", id), nil)
	if err != nil {
		return nil, err
	}

	var result User
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// UpdateUser updates an existing user's role
// Note: According to n8n API docs, only the role can be updated via PATCH /users/{id}/role
func (c *Client) UpdateUser(id string, user *User) (*User, error) {
	// Update the role if it's provided
	if user.Role != "" {
		type UpdateRoleRequest struct {
			NewRoleName string `json:"newRoleName"`
		}

		request := UpdateRoleRequest{
			NewRoleName: user.Role,
		}

		_, err := c.doRequest("PATCH", fmt.Sprintf("/api/v1/users/%s/role", id), request)
		if err != nil {
			return nil, err
		}
	}

	// After updating, fetch the user to get the current state
	updatedUser, err := c.GetUser(id)
	if err != nil {
		return nil, err
	}

	// If the API doesn't return the role in GetUser response, preserve the role from the request
	// This handles cases where n8n API doesn't return role/globalRole in the GET response
	if updatedUser.GetRole() == "" && user.Role != "" {
		updatedUser.SetRole(user.Role)
	}

	return updatedUser, nil
}

// DeleteUser deletes a user
func (c *Client) DeleteUser(id string) error {
	_, err := c.doRequest("DELETE", fmt.Sprintf("/api/v1/users/%s", id), nil)
	return err
}

// ListUsers lists all users
func (c *Client) ListUsers() ([]User, error) {
	respBody, err := c.doRequest("GET", "/api/v1/users", nil)
	if err != nil {
		return nil, err
	}

	var result UserListResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.Data, nil
}
