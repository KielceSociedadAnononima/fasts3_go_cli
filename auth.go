package fasts3

import (
	"context"
	"net/http"
)

// VerifyAuth validates the currently configured credentials (`GET /api/auth/verify`).
func (c *Client) VerifyAuth(ctx context.Context) (*AuthVerifyResponse, error) {
	var out AuthVerifyResponse
	if err := c.doJSON(ctx, http.MethodGet, "/api/auth/verify", nil, nil, &out, true); err != nil {
		return nil, err
	}
	return &out, nil
}

// BootstrapStatus checks whether initial server bootstrap configuration has been completed (`GET /api/auth/bootstrap-status`).
func (c *Client) BootstrapStatus(ctx context.Context) (*BootstrapStatusResponse, error) {
	var out BootstrapStatusResponse
	if err := c.doJSON(ctx, http.MethodGet, "/api/auth/bootstrap-status", nil, nil, &out, false); err != nil {
		return nil, err
	}
	return &out, nil
}

// BootstrapLogin validates a Bootstrap Token via `POST /api/auth/bootstrap-login` without mutating client credentials.
func (c *Client) BootstrapLogin(ctx context.Context, token string) (*BootstrapLoginResponse, error) {
	req := BootstrapLoginRequest{Token: token}
	var out BootstrapLoginResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/auth/bootstrap-login", nil, req, &out, false); err != nil {
		return nil, err
	}
	return &out, nil
}

// BootstrapLoginAndAuthenticate validates a Bootstrap Token and stores it as Bearer credentials on the Client.
func (c *Client) BootstrapLoginAndAuthenticate(ctx context.Context, token string) (*BootstrapLoginResponse, error) {
	resp, err := c.BootstrapLogin(ctx, token)
	if err != nil {
		return nil, err
	}
	c.SetCredentials(BearerCredentials(resp.Token))
	return resp, nil
}

// BootstrapSetup configures the initial administrator user using a Bootstrap Token (`POST /api/auth/bootstrap-setup`).
func (c *Client) BootstrapSetup(ctx context.Context, req BootstrapSetupRequest) (*BootstrapSetupResponse, error) {
	var out BootstrapSetupResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/auth/bootstrap-setup", nil, req, &out, false); err != nil {
		return nil, err
	}
	return &out, nil
}

// BootstrapSetupAndAuthenticate runs BootstrapSetup and updates the Client with the new Basic Auth credentials.
func (c *Client) BootstrapSetupAndAuthenticate(ctx context.Context, req BootstrapSetupRequest) (*BootstrapSetupResponse, error) {
	resp, err := c.BootstrapSetup(ctx, req)
	if err != nil {
		return nil, err
	}
	if creds, ok := CredentialsFromBasicHeader(resp.BasicAuth); ok {
		c.SetCredentials(creds)
	} else {
		c.SetCredentials(BasicCredentials(req.Username, req.Password))
	}
	return resp, nil
}

// Login authenticates with username and password (`POST /api/auth/login`) without mutating client credentials.
func (c *Client) Login(ctx context.Context, username, password string) (*UserLoginResponse, error) {
	req := UserLoginRequest{Username: username, Password: password}
	var out UserLoginResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/auth/login", nil, req, &out, false); err != nil {
		return nil, err
	}
	return &out, nil
}

// LoginAndAuthenticate authenticates via Login and stores the resulting Basic Auth credentials on the Client.
func (c *Client) LoginAndAuthenticate(ctx context.Context, username, password string) (*UserLoginResponse, error) {
	resp, err := c.Login(ctx, username, password)
	if err != nil {
		return nil, err
	}
	if creds, ok := CredentialsFromBasicHeader(resp.BasicAuth); ok {
		c.SetCredentials(creds)
	} else {
		c.SetCredentials(BasicCredentials(username, password))
	}
	return resp, nil
}

// ChangePassword updates the currently authenticated user's password (`POST /api/auth/change-password`).
func (c *Client) ChangePassword(ctx context.Context, req ChangePasswordRequest) (*ChangePasswordResponse, error) {
	var out ChangePasswordResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/auth/change-password", nil, req, &out, true); err != nil {
		return nil, err
	}
	return &out, nil
}

// ChangePasswordAndAuthenticate updates the active user's password and rotates stored Basic Auth credentials.
func (c *Client) ChangePasswordAndAuthenticate(ctx context.Context, req ChangePasswordRequest) (*ChangePasswordResponse, error) {
	resp, err := c.ChangePassword(ctx, req)
	if err != nil {
		return nil, err
	}
	if creds, ok := CredentialsFromBasicHeader(resp.BasicAuth); ok {
		c.SetCredentials(creds)
	}
	return resp, nil
}

// GetBootstrapToken retrieves the server's Bootstrap Token (`GET /api/auth/bootstrap-token`, requires Admin).
func (c *Client) GetBootstrapToken(ctx context.Context) (*string, error) {
	var out BootstrapTokenResponse
	if err := c.doJSON(ctx, http.MethodGet, "/api/auth/bootstrap-token", nil, nil, &out, true); err != nil {
		return nil, err
	}
	return out.BootstrapToken, nil
}
