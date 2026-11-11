package fasts3

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func buildFileListQuery(filter *FileListFilter) url.Values {
	if filter == nil {
		return nil
	}
	q := url.Values{}
	if filter.Name != "" {
		q.Set("name", filter.Name)
	}
	if filter.Subpath != "" {
		q.Set("subpath", filter.Subpath)
	}
	return q
}

// ListFiles lists object keys in bucket with optional Name / Subpath filtering (`GET /api/buckets/{bucket}/files`).
func (c *Client) ListFiles(ctx context.Context, bucket string, filter *FileListFilter) ([]string, error) {
	var out []string
	path := fmt.Sprintf("/api/buckets/%s/files", encodePathSegment(bucket))
	if err := c.doJSON(ctx, http.MethodGet, path, buildFileListQuery(filter), nil, &out, true); err != nil {
		return nil, err
	}
	return out, nil
}

// ListFileDetails lists rich file metadata in bucket (`GET /api/buckets/{bucket}/files/details`).
func (c *Client) ListFileDetails(ctx context.Context, bucket string, filter *FileListFilter) ([]FileDetail, error) {
	var out []FileDetail
	path := fmt.Sprintf("/api/buckets/%s/files/details", encodePathSegment(bucket))
	if err := c.doJSON(ctx, http.MethodGet, path, buildFileListQuery(filter), nil, &out, true); err != nil {
		return nil, err
	}
	return out, nil
}

// UploadBytes uploads in-memory bytes via `PUT /api/buckets/{bucket}/files/{key}`.
func (c *Client) UploadBytes(ctx context.Context, bucket, key string, data []byte, contentType string) (*StatusMessageResponse, error) {
	return c.UploadStream(ctx, bucket, key, bytes.NewReader(data), int64(len(data)), contentType)
}

// UploadStream streams an io.Reader directly to `PUT /api/buckets/{bucket}/files/{key}` without buffering the payload in memory.
// Pass contentLength >= 0 to set an explicit Content-Length header, or -1 if unknown.
func (c *Client) UploadStream(ctx context.Context, bucket, key string, reader io.Reader, contentLength int64, contentType string) (*StatusMessageResponse, error) {
	encodedKey, err := encodeObjectKey(key)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/api/buckets/%s/files/%s", encodePathSegment(bucket), encodedKey)

	req, err := c.newRequest(ctx, http.MethodPut, path, nil, reader, true)
	if err != nil {
		return nil, err
	}
	if contentLength >= 0 {
		req.ContentLength = contentLength
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkResponseError(resp); err != nil {
		return nil, err
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var out StatusMessageResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("fasts3: failed to decode upload response: %w", err)
	}
	return &out, nil
}

// UploadFile streams a local file from disk to `PUT /api/buckets/{bucket}/files/{key}`.
func (c *Client) UploadFile(ctx context.Context, bucket, key, localPath string) (*StatusMessageResponse, error) {
	f, err := os.Open(localPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		return nil, fmt.Errorf("fasts3: %q is a directory, not a file", localPath)
	}

	return c.UploadStream(ctx, bucket, key, f, stat.Size(), "")
}

// UploadJSON uploads a Base64-encoded file payload via `POST /api/buckets/{bucket}/files/upload`.
func (c *Client) UploadJSON(ctx context.Context, bucket string, req JSONUploadRequest) (*StatusMessageResponse, error) {
	var out StatusMessageResponse
	path := fmt.Sprintf("/api/buckets/%s/files/upload", encodePathSegment(bucket))
	if err := c.doJSON(ctx, http.MethodPost, path, nil, req, &out, true); err != nil {
		return nil, err
	}
	return &out, nil
}

// UploadJSONBytes Base64-encodes data and uploads it via `POST /api/buckets/{bucket}/files/upload`.
func (c *Client) UploadJSONBytes(ctx context.Context, bucket, key string, name *string, data []byte) (*StatusMessageResponse, error) {
	return c.UploadJSON(ctx, bucket, JSONUploadRequest{
		Name: name,
		Key:  key,
		Data: base64.StdEncoding.EncodeToString(data),
	})
}

// DownloadStream opens a streaming HTTP download (`GET /api/buckets/{bucket}/files/{key}`).
// The caller MUST close DownloadStream.Body when finished reading.
func (c *Client) DownloadStream(ctx context.Context, bucket, key string, opts *DownloadOptions) (*DownloadStream, error) {
	encodedKey, err := encodeObjectKey(key)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/api/buckets/%s/files/%s", encodePathSegment(bucket), encodedKey)

	var q url.Values
	if opts != nil {
		q = url.Values{}
		if opts.Thumbnail {
			q.Set("thumbnail", "true")
		}
		if opts.Width > 0 {
			q.Set("width", strconv.FormatUint(uint64(opts.Width), 10))
		}
		if opts.Height > 0 {
			q.Set("height", strconv.FormatUint(uint64(opts.Height), 10))
		}
	}

	req, err := c.newRequest(ctx, http.MethodGet, path, q, nil, true)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if err := checkResponseError(resp); err != nil {
		_ = resp.Body.Close()
		return nil, err
	}

	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/octet-stream"
	}

	return &DownloadStream{
		Body:          resp.Body,
		ContentType:   ct,
		ContentLength: resp.ContentLength,
		Filename:      parseContentDispositionFilename(resp.Header.Get("Content-Disposition")),
	}, nil
}

// DownloadWithOptions downloads an object into memory with optional thumbnail query options.
func (c *Client) DownloadWithOptions(ctx context.Context, bucket, key string, opts *DownloadOptions) (*DownloadedFile, error) {
	stream, err := c.DownloadStream(ctx, bucket, key, opts)
	if err != nil {
		return nil, err
	}
	defer stream.Body.Close()

	data, err := io.ReadAll(stream.Body)
	if err != nil {
		return nil, err
	}

	return &DownloadedFile{
		Data:          data,
		ContentType:   stream.ContentType,
		ContentLength: stream.ContentLength,
		Filename:      stream.Filename,
	}, nil
}

// DownloadFile downloads an object and its HTTP headers (`GET /api/buckets/{bucket}/files/{key}`).
func (c *Client) DownloadFile(ctx context.Context, bucket, key string) (*DownloadedFile, error) {
	return c.DownloadWithOptions(ctx, bucket, key, nil)
}

// DownloadBytes downloads raw object bytes (`GET /api/buckets/{bucket}/files/{key}`).
func (c *Client) DownloadBytes(ctx context.Context, bucket, key string) ([]byte, error) {
	dl, err := c.DownloadFile(ctx, bucket, key)
	if err != nil {
		return nil, err
	}
	return dl.Data, nil
}

// DownloadToWriter streams an object directly into w without buffering the whole payload in RAM.
func (c *Client) DownloadToWriter(ctx context.Context, bucket, key string, w io.Writer) (int64, error) {
	stream, err := c.DownloadStream(ctx, bucket, key, nil)
	if err != nil {
		return 0, err
	}
	defer stream.Body.Close()

	return io.Copy(w, stream.Body)
}

// DownloadToPath streams an object directly to destPath on disk, creating parent directories if needed.
func (c *Client) DownloadToPath(ctx context.Context, bucket, key, destPath string) (int64, error) {
	if dir := filepath.Dir(destPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return 0, err
		}
	}
	f, err := os.Create(destPath)
	if err != nil {
		return 0, err
	}
	written, copyErr := c.DownloadToWriter(ctx, bucket, key, f)
	closeErr := f.Close()
	if copyErr != nil {
		_ = os.Remove(destPath)
		return written, copyErr
	}
	return written, closeErr
}

// GetThumbnail downloads a resized image thumbnail via `GET /api/buckets/{bucket}/files/{key}/thumbnail`.
func (c *Client) GetThumbnail(ctx context.Context, bucket, key string, opts *ThumbnailOptions) (*ThumbnailResponse, error) {
	encodedKey, err := encodeObjectKey(key)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/api/buckets/%s/files/%s/thumbnail", encodePathSegment(bucket), encodedKey)
	return c.fetchThumbnailRoute(ctx, path, opts)
}

// GetThumbnailAlt downloads a resized image thumbnail via `GET /api/buckets/{bucket}/thumbnail/{key}`.
func (c *Client) GetThumbnailAlt(ctx context.Context, bucket, key string, opts *ThumbnailOptions) (*ThumbnailResponse, error) {
	encodedKey, err := encodeObjectKey(key)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/api/buckets/%s/thumbnail/%s", encodePathSegment(bucket), encodedKey)
	return c.fetchThumbnailRoute(ctx, path, opts)
}

func (c *Client) fetchThumbnailRoute(ctx context.Context, path string, opts *ThumbnailOptions) (*ThumbnailResponse, error) {
	var q url.Values
	if opts != nil {
		q = url.Values{}
		if opts.Width > 0 {
			q.Set("width", strconv.FormatUint(uint64(opts.Width), 10))
		}
		if opts.Height > 0 {
			q.Set("height", strconv.FormatUint(uint64(opts.Height), 10))
		}
	}

	req, err := c.newRequest(ctx, http.MethodGet, path, q, nil, true)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := checkResponseError(resp); err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "image/png"
	}
	return &ThumbnailResponse{
		Data:        data,
		ContentType: ct,
	}, nil
}

// SaveThumbnailToPath fetches a thumbnail and writes it to destPath on disk.
func (c *Client) SaveThumbnailToPath(ctx context.Context, bucket, key string, opts *ThumbnailOptions, destPath string) error {
	thumb, err := c.GetThumbnail(ctx, bucket, key, opts)
	if err != nil {
		return err
	}
	if dir := filepath.Dir(destPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(destPath, thumb.Data, 0o644)
}

// DeleteFile deletes an object from a bucket (`DELETE /api/buckets/{bucket}/files/{key}`).
func (c *Client) DeleteFile(ctx context.Context, bucket, key string) (*StatusMessageResponse, error) {
	encodedKey, err := encodeObjectKey(key)
	if err != nil {
		return nil, err
	}
	var out StatusMessageResponse
	path := fmt.Sprintf("/api/buckets/%s/files/%s", encodePathSegment(bucket), encodedKey)
	if err := c.doJSON(ctx, http.MethodDelete, path, nil, nil, &out, true); err != nil {
		return nil, err
	}
	return &out, nil
}

// UploadDirectory recursively uploads all files inside localDir into bucket under optional keyPrefix.
// Returns the sorted list of uploaded S3 object keys.
func (c *Client) UploadDirectory(ctx context.Context, bucket, keyPrefix, localDir string) ([]string, error) {
	stat, err := os.Stat(localDir)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("fasts3: %q is not a directory", localDir)
	}

	cleanPrefix := strings.Trim(strings.ReplaceAll(keyPrefix, "\\", "/"), "/")
	var uploaded []string

	err = filepath.WalkDir(localDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		rel, err := filepath.Rel(localDir, path)
		if err != nil {
			return err
		}
		relSlash := filepath.ToSlash(rel)
		fullKey := relSlash
		if cleanPrefix != "" {
			fullKey = cleanPrefix + "/" + relSlash
		}

		if _, err := c.UploadFile(ctx, bucket, fullKey, path); err != nil {
			return err
		}
		uploaded = append(uploaded, fullKey)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(uploaded)
	return uploaded, nil
}

// DownloadPrefixToDir downloads all objects matching subpath (or all objects when subpath == "") into destDir,
// safely preserving relative subdirectory hierarchy. Returns the list of written local file paths.
func (c *Client) DownloadPrefixToDir(ctx context.Context, bucket, subpath, destDir string) ([]string, error) {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return nil, err
	}
	cleanDest, err := filepath.Abs(destDir)
	if err != nil {
		return nil, err
	}

	var filter *FileListFilter
	if subpath != "" {
		filter = &FileListFilter{Subpath: subpath}
	}
	keys, err := c.ListFiles(ctx, bucket, filter)
	if err != nil {
		return nil, err
	}

	written := make([]string, 0, len(keys))
	for _, key := range keys {
		var parts []string
		for _, seg := range strings.Split(strings.ReplaceAll(key, "\\", "/"), "/") {
			if seg == "" || seg == "." || seg == ".." {
				continue
			}
			parts = append(parts, seg)
		}
		if len(parts) == 0 {
			continue
		}
		target := filepath.Join(append([]string{cleanDest}, parts...)...)
		relCheck, err := filepath.Rel(cleanDest, target)
		if err != nil || strings.HasPrefix(relCheck, "..") {
			continue
		}
		if _, err := c.DownloadToPath(ctx, bucket, key, target); err != nil {
			return written, err
		}
		written = append(written, target)
	}
	return written, nil
}

// DeletePrefix deletes all objects under subpath in bucket, returning the count of deleted keys.
func (c *Client) DeletePrefix(ctx context.Context, bucket, subpath string) (int, error) {
	keys, err := c.ListFiles(ctx, bucket, &FileListFilter{Subpath: subpath})
	if err != nil {
		return 0, err
	}
	for _, k := range keys {
		if _, err := c.DeleteFile(ctx, bucket, k); err != nil {
			return 0, err
		}
	}
	return len(keys), nil
}

// FileURL builds a direct HTTP URL for `GET /api/buckets/{bucket}/files/{key}`.
// When includeAuthQuery is true, appends FastS3's query authentication parameter (`?token=...`, `?api_key=...`, or `?auth=...`).
func (c *Client) FileURL(bucket, key string, includeAuthQuery bool) (string, error) {
	encodedKey, err := encodeObjectKey(key)
	if err != nil {
		return "", err
	}
	base := fmt.Sprintf("%s/api/buckets/%s/files/%s", c.baseURL, encodePathSegment(bucket), encodedKey)
	if includeAuthQuery {
		if qk, qv, ok := c.Credentials().toQueryParam(); ok {
			return fmt.Sprintf("%s?%s=%s", base, qk, qv), nil
		}
	}
	return base, nil
}

// ThumbnailURL builds a direct HTTP URL for `GET /api/buckets/{bucket}/files/{key}/thumbnail`.
func (c *Client) ThumbnailURL(bucket, key string, opts *ThumbnailOptions, includeAuthQuery bool) (string, error) {
	encodedKey, err := encodeObjectKey(key)
	if err != nil {
		return "", err
	}
	base := fmt.Sprintf("%s/api/buckets/%s/files/%s/thumbnail", c.baseURL, encodePathSegment(bucket), encodedKey)

	var params []string
	if opts != nil {
		if opts.Width > 0 {
			params = append(params, fmt.Sprintf("width=%d", opts.Width))
		}
		if opts.Height > 0 {
			params = append(params, fmt.Sprintf("height=%d", opts.Height))
		}
	}
	if includeAuthQuery {
		if qk, qv, ok := c.Credentials().toQueryParam(); ok {
			params = append(params, fmt.Sprintf("%s=%s", qk, qv))
		}
	}

	if len(params) == 0 {
		return base, nil
	}
	return base + "?" + strings.Join(params, "&"), nil
}
