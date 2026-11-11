package fasts3

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// apiErrorPayload mirrors the JSON error envelope returned by FastS3 (`{"error": "...", "status": 404}`).
type apiErrorPayload struct {
	Error  string `json:"error"`
	Status int    `json:"status"`
}

// APIError represents a non-2xx HTTP error response returned by the FastS3 server.
type APIError struct {
	StatusCode int
	Message    string
}

// Error implements the standard error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("fasts3 api error (HTTP %d): %s", e.StatusCode, e.Message)
}

// IsBadRequest reports whether the HTTP status code is 400.
func (e *APIError) IsBadRequest() bool {
	return e.StatusCode == http.StatusBadRequest
}

// IsUnauthorized reports whether the HTTP status code is 401.
func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

// IsForbidden reports whether the HTTP status code is 403.
func (e *APIError) IsForbidden() bool {
	return e.StatusCode == http.StatusForbidden
}

// IsNotFound reports whether the HTTP status code is 404.
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsConflict reports whether the HTTP status code is 409.
func (e *APIError) IsConflict() bool {
	return e.StatusCode == http.StatusConflict
}

// IsQuotaExceeded reports whether the error represents a bucket storage quota violation (HTTP 403 with quota message).
func (e *APIError) IsQuotaExceeded() bool {
	return e.StatusCode == http.StatusForbidden && strings.Contains(strings.ToLower(e.Message), "quota")
}

// IsBadRequest checks if any error in err's chain is a 400 *APIError.
func IsBadRequest(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.IsBadRequest()
}

// IsUnauthorized checks if any error in err's chain is a 401 *APIError.
func IsUnauthorized(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.IsUnauthorized()
}

// IsForbidden checks if any error in err's chain is a 403 *APIError.
func IsForbidden(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.IsForbidden()
}

// IsNotFound checks if any error in err's chain is a 404 *APIError.
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.IsNotFound()
}

// IsConflict checks if any error in err's chain is a 409 *APIError.
func IsConflict(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.IsConflict()
}

// IsQuotaExceeded checks if any error in err's chain indicates bucket quota exhaustion.
func IsQuotaExceeded(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.IsQuotaExceeded()
}
