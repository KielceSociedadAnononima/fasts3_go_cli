package fasts3

import (
	"context"
	"fmt"
	"net/http"
)

// ============================================================================
// API Keys Endpoints
// ============================================================================

// ListAPIKeys lists API keys (`GET /api/api-keys`).
// Admins receive all keys; non-admin users receive only their own keys.
func (c *Client) ListAPIKeys(ctx context.Context) ([]APIKeyInfo, error) {
	var out []APIKeyInfo
	if err := c.doJSON(ctx, http.MethodGet, "/api/api-keys", nil, nil, &out, true); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateAPIKey generates a new user-scoped API key (`POST /api/api-keys`).
func (c *Client) CreateAPIKey(ctx context.Context, req CreateAPIKeyRequest) (*CreatedAPIKey, error) {
	var out CreatedAPIKey
	if err := c.doJSON(ctx, http.MethodPost, "/api/api-keys", nil, req, &out, true); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteAPIKey revokes an API key by its numeric ID (`DELETE /api/api-keys/{id}`).
func (c *Client) DeleteAPIKey(ctx context.Context, id int64) (*StatusMessageResponse, error) {
	var out StatusMessageResponse
	path := fmt.Sprintf("/api/api-keys/%d", id)
	if err := c.doJSON(ctx, http.MethodDelete, path, nil, nil, &out, true); err != nil {
		return nil, err
	}
	return &out, nil
}

// ============================================================================
// Users Endpoints
// ============================================================================

// ListUsers lists all SQLite users (`GET /api/users`, requires Admin).
func (c *Client) ListUsers(ctx context.Context) ([]User, error) {
	var out []User
	if err := c.doJSON(ctx, http.MethodGet, "/api/users", nil, nil, &out, true); err != nil {
		return nil, err
	}
	return out, nil
}

// GetUserByUsername finds a user by exact username via ListUsers, returning (nil, nil) if not found.
func (c *Client) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	users, err := c.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	for i := range users {
		if users[i].Username == username {
			u := users[i]
			return &u, nil
		}
	}
	return nil, nil
}

// CreateUser creates a new user (`POST /api/users`, requires Admin).
func (c *Client) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
	var out User
	if err := c.doJSON(ctx, http.MethodPost, "/api/users", nil, req, &out, true); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateUser updates a user's password and/or admin privileges (`PUT /api/users/{id}`).
func (c *Client) UpdateUser(ctx context.Context, id int64, req UpdateUserRequest) (*User, error) {
	var out User
	path := fmt.Sprintf("/api/users/%d", id)
	if err := c.doJSON(ctx, http.MethodPut, path, nil, req, &out, true); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteUser deletes a user by numeric ID (`DELETE /api/users/{id}`, requires Admin).
func (c *Client) DeleteUser(ctx context.Context, id int64) (*StatusMessageResponse, error) {
	var out StatusMessageResponse
	path := fmt.Sprintf("/api/users/%d", id)
	if err := c.doJSON(ctx, http.MethodDelete, path, nil, nil, &out, true); err != nil {
		return nil, err
	}
	return &out, nil
}

// ============================================================================
// Permissions Endpoints
// ============================================================================

// ListPermissions lists all bucket permission assignments (`GET /api/permissions`, requires Admin).
func (c *Client) ListPermissions(ctx context.Context) ([]Permission, error) {
	var out []Permission
	if err := c.doJSON(ctx, http.MethodGet, "/api/permissions", nil, nil, &out, true); err != nil {
		return nil, err
	}
	return out, nil
}

// SetPermission grants or updates user permissions on a bucket (`POST /api/permissions`, requires Admin).
func (c *Client) SetPermission(ctx context.Context, req SetPermissionRequest) (*Permission, error) {
	var out Permission
	if err := c.doJSON(ctx, http.MethodPost, "/api/permissions", nil, req, &out, true); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeletePermission revokes a permission entry by numeric permission ID (`DELETE /api/permissions/{id}`, requires Admin).
func (c *Client) DeletePermission(ctx context.Context, id int64) (*StatusMessageResponse, error) {
	var out StatusMessageResponse
	path := fmt.Sprintf("/api/permissions/%d", id)
	if err := c.doJSON(ctx, http.MethodDelete, path, nil, nil, &out, true); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetBucketPermissions lists all user permissions configured on a bucket (`GET /api/buckets/{bucket}/permissions`).
func (c *Client) GetBucketPermissions(ctx context.Context, bucket string) ([]Permission, error) {
	var out []Permission
	path := fmt.Sprintf("/api/buckets/%s/permissions", encodePathSegment(bucket))
	if err := c.doJSON(ctx, http.MethodGet, path, nil, nil, &out, true); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteBucketPermission revokes a specific user's permissions on a bucket (`DELETE /api/buckets/{bucket}/permissions/{user_id}`, requires Admin).
func (c *Client) DeleteBucketPermission(ctx context.Context, bucket string, userID int64) (*StatusMessageResponse, error) {
	var out StatusMessageResponse
	path := fmt.Sprintf("/api/buckets/%s/permissions/%d", encodePathSegment(bucket), userID)
	if err := c.doJSON(ctx, http.MethodDelete, path, nil, nil, &out, true); err != nil {
		return nil, err
	}
	return &out, nil
}
