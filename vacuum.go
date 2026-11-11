package fasts3

import (
	"context"
	"net/http"
)

// VacuumStatus retrieves metrics from the latest AutoVacuum execution (`GET /api/vacuum/status`, requires Admin).
func (c *Client) VacuumStatus(ctx context.Context) (*VacuumStats, error) {
	var out VacuumStats
	if err := c.doJSON(ctx, http.MethodGet, "/api/vacuum/status", nil, nil, &out, true); err != nil {
		return nil, err
	}
	return &out, nil
}

// TriggerVacuum immediately runs an AutoVacuum cleanup and bucket quota reconciliation cycle (`POST /api/vacuum`, requires Admin).
func (c *Client) TriggerVacuum(ctx context.Context) (*VacuumStats, error) {
	var out VacuumStats
	if err := c.doJSON(ctx, http.MethodPost, "/api/vacuum", nil, nil, &out, true); err != nil {
		return nil, err
	}
	return &out, nil
}
