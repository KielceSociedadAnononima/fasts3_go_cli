package fasts3

import "io"

// StatusMessageResponse represents the standard `{"status": "...", "message": "..."}` response returned by mutating endpoints.
type StatusMessageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ============================================================================
// Authentication & Bootstrap Models
// ============================================================================

// AuthVerifyResponse is returned by `GET /api/auth/verify`.
type AuthVerifyResponse struct {
	Status      string       `json:"status"`
	IsAdmin     bool         `json:"is_admin"`
	Username    string       `json:"username"`
	UserID      *int64       `json:"user_id,omitempty"`
	IsBootstrap *bool        `json:"is_bootstrap,omitempty"`
	Permissions []Permission `json:"permissions,omitempty"`
}

// BootstrapStatusResponse is returned by `GET /api/auth/bootstrap-status`.
type BootstrapStatusResponse struct {
	IsConfigured bool   `json:"is_configured"`
	Message      string `json:"message"`
}

// BootstrapLoginRequest is the payload for `POST /api/auth/bootstrap-login`.
type BootstrapLoginRequest struct {
	Token string `json:"token"`
}

// BootstrapLoginResponse is returned by `POST /api/auth/bootstrap-login`.
type BootstrapLoginResponse struct {
	Status   string `json:"status"`
	IsAdmin  bool   `json:"is_admin"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

// BootstrapSetupRequest is the payload for `POST /api/auth/bootstrap-setup`.
type BootstrapSetupRequest struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// BootstrapSetupResponse is returned by `POST /api/auth/bootstrap-setup`.
type BootstrapSetupResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Username  string `json:"username"`
	UserID    int64  `json:"user_id"`
	IsAdmin   bool   `json:"is_admin"`
	BasicAuth string `json:"basic_auth"`
}

// UserLoginRequest is the payload for `POST /api/auth/login`.
type UserLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserLoginResponse is returned by `POST /api/auth/login`.
type UserLoginResponse struct {
	Status    string `json:"status"`
	IsAdmin   bool   `json:"is_admin"`
	Username  string `json:"username"`
	UserID    *int64 `json:"user_id,omitempty"`
	BasicAuth string `json:"basic_auth"`
}

// ChangePasswordRequest is the payload for `POST /api/auth/change-password`.
type ChangePasswordRequest struct {
	CurrentPassword *string `json:"current_password,omitempty"`
	NewPassword     string  `json:"new_password"`
}

// ChangePasswordResponse is returned by `POST /api/auth/change-password`.
type ChangePasswordResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Username  string `json:"username"`
	BasicAuth string `json:"basic_auth"`
}

// BootstrapTokenResponse is returned by `GET /api/auth/bootstrap-token`.
type BootstrapTokenResponse struct {
	BootstrapToken *string `json:"bootstrap_token"`
}

// ============================================================================
// API Keys Models
// ============================================================================

// APIKeyInfo describes a stored API key returned by `GET /api/api-keys`.
type APIKeyInfo struct {
	ID         int64   `json:"id"`
	UserID     int64   `json:"user_id"`
	Username   string  `json:"username"`
	IsAdmin    bool    `json:"is_admin"`
	Name       string  `json:"name"`
	KeyPrefix  string  `json:"key_prefix"`
	CreatedAt  string  `json:"created_at"`
	LastUsedAt *string `json:"last_used_at"`
}

// CreatedAPIKey is returned once upon creating an API key via `POST /api/api-keys`.
type CreatedAPIKey struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	IsAdmin   bool   `json:"is_admin"`
	Name      string `json:"name"`
	KeyPrefix string `json:"key_prefix"`
	APIKey    string `json:"api_key"`
	CreatedAt string `json:"created_at"`
}

// CreateAPIKeyRequest is the payload for `POST /api/api-keys`.
type CreateAPIKeyRequest struct {
	Name   string `json:"name"`
	UserID *int64 `json:"user_id,omitempty"`
}

// ============================================================================
// Users & Permissions Models
// ============================================================================

// User represents a user record returned by FastS3 user endpoints.
type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	IsAdmin   bool   `json:"is_admin"`
	CreatedAt string `json:"created_at"`
}

// CreateUserRequest is the payload for `POST /api/users`.
type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	IsAdmin  *bool  `json:"is_admin,omitempty"`
}

// UpdateUserRequest is the payload for `PUT /api/users/{id}`.
type UpdateUserRequest struct {
	Password *string `json:"password,omitempty"`
	IsAdmin  *bool   `json:"is_admin,omitempty"`
}

// Permission represents a user's granular access rights over a bucket.
type Permission struct {
	ID         int64   `json:"id"`
	UserID     int64   `json:"user_id"`
	Username   *string `json:"username,omitempty"`
	BucketName string  `json:"bucket_name"`
	CanRead    bool    `json:"can_read"`
	CanWrite   bool    `json:"can_write"`
	CanDelete  bool    `json:"can_delete"`
}

// SetPermissionRequest is the payload for `POST /api/permissions`.
type SetPermissionRequest struct {
	UserID     int64  `json:"user_id"`
	BucketName string `json:"bucket_name"`
	CanRead    bool   `json:"can_read"`
	CanWrite   bool   `json:"can_write"`
	CanDelete  bool   `json:"can_delete"`
}

// ============================================================================
// Multi-Disk Storage Models
// ============================================================================

// Disk represents a physical storage volume registered in FastS3.
type Disk struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Path            string  `json:"path"`
	IsDefault       bool    `json:"is_default"`
	CreatedAt       string  `json:"created_at"`
	TotalSpace      *uint64 `json:"total_space,omitempty"`
	AvailableSpace  *uint64 `json:"available_space,omitempty"`
	UsedSpace       *uint64 `json:"used_space,omitempty"`
	FastS3UsedSpace *uint64 `json:"fasts3_used_space,omitempty"`
}

// CreateDiskRequest is the payload for `POST /api/disks`.
type CreateDiskRequest struct {
	ID   *string `json:"id,omitempty"`
	Name string  `json:"name"`
	Path string  `json:"path"`
}

// ============================================================================
// Bucket Models
// ============================================================================

// Bucket represents a FastS3 storage bucket.
type Bucket struct {
	ID       uint64  `json:"id"`
	Name     string  `json:"name"`
	IsPublic bool    `json:"is_public"`
	DiskPath *string `json:"disk_path"`
	DiskID   *string `json:"disk_id"`
	Quota    uint64  `json:"quota"`
	Used     uint64  `json:"used"`
}

// CreateBucketRequest is the payload for `POST /api/buckets`.
type CreateBucketRequest struct {
	Name     string  `json:"name"`
	IsPublic bool    `json:"is_public"`
	DiskID   *string `json:"disk_id,omitempty"`
	DiskPath *string `json:"disk_path,omitempty"`
	Quota    uint64  `json:"quota"`
}

// UpdateBucketRequest is the payload for `PUT /api/buckets/{id}`.
// Note: FastS3 expects `name` to be present in the JSON body.
type UpdateBucketRequest struct {
	Name     string  `json:"name"`
	IsPublic bool    `json:"is_public"`
	DiskID   *string `json:"disk_id,omitempty"`
	DiskPath *string `json:"disk_path,omitempty"`
	Quota    uint64  `json:"quota"`
}

// ============================================================================
// Files & Thumbnails Models
// ============================================================================

// FileListFilter holds optional query filters for `ListFiles` and `ListFileDetails`.
type FileListFilter struct {
	// Name filters filenames by case-insensitive substring.
	Name string
	// Subpath restricts results to keys under the given folder prefix.
	Subpath string
}

// FileDetail contains rich file metadata returned by `GET /api/buckets/{bucket}/files/details`.
type FileDetail struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	Size      uint64 `json:"size"`
	MimeType  string `json:"mime_type"`
	Extension string `json:"extension"`
	IsImage   bool   `json:"is_image"`
}

// JSONUploadRequest is the payload for `POST /api/buckets/{bucket}/files/upload`.
type JSONUploadRequest struct {
	Name *string `json:"name,omitempty"`
	Key  string  `json:"key"`
	// Data must be Base64-encoded file bytes.
	Data string `json:"data"`
}

// ThumbnailOptions specifies target resize dimensions for thumbnail generation.
type ThumbnailOptions struct {
	Width  uint32
	Height uint32
}

// DownloadOptions configures query options on `GET /api/buckets/{bucket}/files/{key}`.
type DownloadOptions struct {
	Thumbnail bool
	Width     uint32
	Height    uint32
}

// DownloadedFile holds an in-memory downloaded object and its HTTP response headers.
type DownloadedFile struct {
	Data          []byte
	ContentType   string
	ContentLength int64
	Filename      string
}

// DownloadStream wraps a streaming response body for large object downloads.
// Callers MUST close Body when finished reading.
type DownloadStream struct {
	Body          io.ReadCloser
	ContentType   string
	ContentLength int64
	Filename      string
}

// ThumbnailResponse holds a generated image thumbnail and its MIME type.
type ThumbnailResponse struct {
	Data        []byte
	ContentType string
}

// ============================================================================
// AutoVacuum Models
// ============================================================================

// VacuumStats contains metrics returned by `GET /api/vacuum/status` and `POST /api/vacuum`.
type VacuumStats struct {
	FilesDeleted      uint64 `json:"files_deleted"`
	BytesReclaimed    uint64 `json:"bytes_reclaimed"`
	BucketsReconciled uint64 `json:"buckets_reconciled"`
	LastRunTimestamp  uint64 `json:"last_run_timestamp"`
	DurationMs        uint64 `json:"duration_ms"`
}

// BoolPtr returns a pointer to b (convenience helper for optional request fields).
func BoolPtr(b bool) *bool { return &b }

// StringPtr returns a pointer to s (convenience helper for optional request fields).
func StringPtr(s string) *string { return &s }

// Int64Ptr returns a pointer to v (convenience helper for optional request fields).
func Int64Ptr(v int64) *int64 { return &v }
