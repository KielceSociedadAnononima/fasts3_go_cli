package fasts3

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const defaultUserAgent = "fasts3-go-client/0.1.0"

// Option configures a Client during construction with NewClient.
type Option func(*Client)

// WithCredentials sets the initial Credentials on the Client.
func WithCredentials(creds Credentials) Option {
	return func(c *Client) {
		c.creds = creds
	}
}

// WithBasicAuth configures HTTP Basic Authentication (`Authorization: Basic ...`).
func WithBasicAuth(username, password string) Option {
	return WithCredentials(BasicCredentials(username, password))
}

// WithCustomHeaderAuth configures `X-FastS3-User` and `X-FastS3-Pass` authentication headers.
func WithCustomHeaderAuth(username, password string) Option {
	return WithCredentials(CustomHeaderCredentials(username, password))
}

// WithBearerToken configures `Authorization: Bearer <token>` (for Bootstrap Token or API Key).
func WithBearerToken(token string) Option {
	return WithCredentials(BearerCredentials(token))
}

// WithAPIKey configures `X-API-Key: <apiKey>` authentication.
func WithAPIKey(apiKey string) Option {
	return WithCredentials(APIKeyCredentials(apiKey))
}

// WithFastS3Token configures `X-FastS3-Token: <token>` authentication.
func WithFastS3Token(token string) Option {
	return WithCredentials(FastS3TokenCredentials(token))
}

// WithHTTPClient supplies a custom *http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// WithTimeout configures the request timeout on the underlying http.Client.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithUserAgent overrides the default User-Agent header sent on outgoing requests.
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		if strings.TrimSpace(ua) != "" {
			c.userAgent = ua
		}
	}
}

// Client is a thread-safe Go client for the FastS3 REST API.
type Client struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string

	mu    sync.RWMutex
	creds Credentials
}

// NewClient constructs a new FastS3 Client pointing at baseURL (e.g. "http://localhost:8080").
func NewClient(baseURL string, opts ...Option) (*Client, error) {
	trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if trimmed == "" {
		return nil, errors.New("fasts3: baseURL cannot be empty")
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return nil, fmt.Errorf("fasts3: invalid baseURL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("fasts3: unsupported URL scheme %q (expected http or https)", parsed.Scheme)
	}
	if parsed.Host == "" {
		return nil, errors.New("fasts3: baseURL is missing host")
	}

	c := &Client{
		baseURL:    trimmed,
		httpClient: &http.Client{Timeout: 5 * time.Minute},
		userAgent:  defaultUserAgent,
		creds:      AnonymousCredentials(),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	return c, nil
}

// BaseURL returns the normalized base URL configured on the Client.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// Credentials returns a snapshot of the currently active Credentials.
func (c *Client) Credentials() Credentials {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.creds
}

// SetCredentials updates the active Credentials safely for all subsequent requests.
func (c *Client) SetCredentials(creds Credentials) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.creds = creds
}

// ClearCredentials resets the client back to anonymous mode.
func (c *Client) ClearCredentials() {
	c.SetCredentials(AnonymousCredentials())
}

func (c *Client) endpointURL(relPath string, query url.Values) string {
	cleanRel := strings.TrimLeft(relPath, "/")
	full := c.baseURL + "/" + cleanRel
	if len(query) > 0 {
		full += "?" + query.Encode()
	}
	return full
}

func (c *Client) newRequest(ctx context.Context, method, relPath string, query url.Values, body io.Reader, applyAuth bool) (*http.Request, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	targetURL := c.endpointURL(relPath, query)
	req, err := http.NewRequestWithContext(ctx, method, targetURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	if applyAuth {
		c.Credentials().apply(req)
	}
	return req, nil
}

func (c *Client) doJSON(ctx context.Context, method, relPath string, query url.Values, reqBody any, respOut any, applyAuth bool) error {
	var bodyReader io.Reader
	if reqBody != nil {
		encoded, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("fasts3: failed to marshal request JSON: %w", err)
		}
		bodyReader = bytes.NewReader(encoded)
	}

	req, err := c.newRequest(ctx, method, relPath, query, bodyReader, applyAuth)
	if err != nil {
		return err
	}
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := checkResponseError(resp); err != nil {
		return err
	}

	if respOut == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, respOut); err != nil {
		return fmt.Errorf("fasts3: failed to decode response JSON: %w", err)
	}
	return nil
}

func checkResponseError(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var payload apiErrorPayload
	if len(raw) > 0 && json.Unmarshal(raw, &payload) == nil && payload.Error != "" {
		status := payload.Status
		if status == 0 {
			status = resp.StatusCode
		}
		return &APIError{
			StatusCode: status,
			Message:    payload.Error,
		}
	}

	msg := strings.TrimSpace(string(raw))
	if msg == "" {
		msg = http.StatusText(resp.StatusCode)
		if msg == "" {
			msg = "unknown HTTP error"
		}
	}
	return &APIError{
		StatusCode: resp.StatusCode,
		Message:    msg,
	}
}
