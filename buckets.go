package fasts3

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// ListBuckets lists buckets accessible to the current credentials (`GET /api/buckets`).
func (c *Client) ListBuckets(ctx context.Context) ([]Bucket, error) {
	var out []Bucket
	if err := c.doJSON(ctx, http.MethodGet, "/api/buckets", nil, nil, &out, true); err != nil {
		return nil, err
	}
	return out, nil
}

// GetBucketByName searches ListBuckets for a bucket matching name (case-insensitive), returning (nil, nil) if not found.
func (c *Client) GetBucketByName(ctx context.Context, name string) (*Bucket, error) {
	buckets, err := c.ListBuckets(ctx)
	if err != nil {
		return nil, err
	}
	for i := range buckets {
		if strings.EqualFold(buckets[i].Name, name) {
			b := buckets[i]
			return &b, nil
		}
	}
	return nil, nil
}

// CreateBucket creates a new storage bucket (`POST /api/buckets`, requires Admin).
func (c *Client) CreateBucket(ctx context.Context, req CreateBucketRequest) (*StatusMessageResponse, error) {
	var out StatusMessageResponse
	if err := c.doJSON(ctx, http.MethodPost, "/api/buckets", nil, req, &out, true); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateBucket updates metadata, quota, or volume assignment of a bucket (`PUT /api/buckets/{id}`, requires Admin).
func (c *Client) UpdateBucket(ctx context.Context, id uint64, req UpdateBucketRequest) (*StatusMessageResponse, error) {
	var out StatusMessageResponse
	path := fmt.Sprintf("/api/buckets/%d", id)
	if err := c.doJSON(ctx, http.MethodPut, path, nil, req, &out, true); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteBucket deletes a bucket and its files by ID and name (`DELETE /api/buckets/{id}/{name}`, requires Admin).
func (c *Client) DeleteBucket(ctx context.Context, id uint64, name string) (*StatusMessageResponse, error) {
	var out StatusMessageResponse
	path := fmt.Sprintf("/api/buckets/%d/%s", id, encodePathSegment(name))
	if err := c.doJSON(ctx, http.MethodDelete, path, nil, nil, &out, true); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteBucketByName resolves the bucket's numeric ID via GetBucketByName and deletes it.
func (c *Client) DeleteBucketByName(ctx context.Context, name string) (*StatusMessageResponse, error) {
	b, err := c.GetBucketByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, &APIError{
			StatusCode: http.StatusNotFound,
			Message:    fmt.Sprintf("Bucket %q not found", name),
		}
	}
	return c.DeleteBucket(ctx, b.ID, b.Name)
}
