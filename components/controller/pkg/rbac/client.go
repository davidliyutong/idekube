package rbac

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents an RBAC client for communicating with the RBAC service
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new RBAC client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CheckPermission checks if a user has permission to perform an action on a resource
func (c *Client) CheckPermission(ctx context.Context, req *CheckPermissionRequest) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/check", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("RBAC service returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var checkResp CheckPermissionResponse
	if err := json.NewDecoder(resp.Body).Decode(&checkResp); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	return checkResp.Allowed, nil
}

// AssignRoleToUser assigns a system role to a user
func (c *Client) AssignRoleToUser(ctx context.Context, userID int64, role string) error {
	url := fmt.Sprintf("%s/api/v1/roles/assign", c.baseURL)

	req := &AssignRoleRequest{
		UserID: userID,
		Role:   role,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("RBAC service returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// AssignRoleForResource assigns a resource-level role to a user
func (c *Client) AssignRoleForResource(ctx context.Context, userID int64, resourceType, resourceID, role string) error {
	url := fmt.Sprintf("%s/api/v1/roles/assign-resource", c.baseURL)

	req := &AssignResourceRoleRequest{
		UserID:       userID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Role:         role,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("RBAC service returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// RevokeRoleForResource revokes a resource-level role from a user
func (c *Client) RevokeRoleForResource(ctx context.Context, userID int64, resourceType, resourceID, role string) error {
	url := fmt.Sprintf("%s/api/v1/roles/revoke-resource", c.baseURL)

	req := &RevokeResourceRoleRequest{
		UserID:       userID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Role:         role,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("RBAC service returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
