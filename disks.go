package fasts3

import (
	"context"
	"fmt"
	"net/http"
)

// ListDisks lists all registered storage volumes and real-time space telemetry (`GET /api/disks`, requires Admin).
func (c *Client) ListDisks(ctx context.Context) ([]Disk, error) {
	var out []Disk
	if err := c.doJSON(ctx, http.MethodGet, "/api/disks", nil, nil, &out, true); err != nil {
		return nil, err
	}
	return out, nil
}

// AddDisk registers a secondary storage volume (`POST /api/disks`, requires Admin).
func (c *Client) AddDisk(ctx context.Context, req CreateDiskRequest) (*Disk, error) {
	var out Disk
	if err := c.doJSON(ctx, http.MethodPost, "/api/disks", nil, req, &out, true); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteDisk unlinks a non-default storage volume with no active buckets (`DELETE /api/disks/{id}`, requires Admin).
func (c *Client) DeleteDisk(ctx context.Context, diskID string) (*StatusMessageResponse, error) {
	var out StatusMessageResponse
	path := fmt.Sprintf("/api/disks/%s", encodePathSegment(diskID))
	if err := c.doJSON(ctx, http.MethodDelete, path, nil, nil, &out, true); err != nil {
		return nil, err
	}
	return &out, nil
}
