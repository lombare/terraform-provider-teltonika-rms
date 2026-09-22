package rms

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// CreditsSummary hits /credits/summary and returns the raw payload.
func (c *Client) CreditsSummary(ctx context.Context) (json.RawMessage, error) {
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := c.Get(ctx, "credits/summary", nil, &env); err != nil {
		return nil, err
	}
	return env.Data, nil
}

// Hotspots lists hotspots across all devices.
func (c *Client) Hotspots(ctx context.Context, q url.Values) ([]Hotspot, error) {
	return ListAll[Hotspot](ctx, c, "hotspots", q)
}

// ListConfiguratorTemplates paginates through /devices/configurator/templates.
func (c *Client) ListConfiguratorTemplates(ctx context.Context, q url.Values) ([]ConfiguratorTemplate, error) {
	return ListAll[ConfiguratorTemplate](ctx, c, "devices/configurator/templates", q)
}

// CreateConfiguratorTemplate posts a raw payload — templates are model-specific and their
// bodies vary wildly, so we let the caller supply the JSON directly.
func (c *Client) CreateConfiguratorTemplate(ctx context.Context, payload json.RawMessage) (string, error) {
	var env struct {
		Data struct {
			ID any `json:"id"`
		} `json:"data"`
	}
	if err := c.Post(ctx, "devices/configurator/templates", payload, &env); err != nil {
		return "", err
	}
	if env.Data.ID == nil {
		return "", fmt.Errorf("rms: configurator template create returned empty id")
	}
	return fmt.Sprint(env.Data.ID), nil
}

func (c *Client) DeleteConfiguratorTemplate(ctx context.Context, id string) error {
	return c.Delete(ctx, "devices/configurator/templates/"+id, nil)
}

// GetConfiguratorTemplate is a best-effort read (the RMS endpoint returns the whole list
// filtered by id via query params — we surface a slice for the caller to pick from).
func (c *Client) GetConfiguratorTemplate(ctx context.Context, id string) (*ConfiguratorTemplate, error) {
	q := url.Values{"id": []string{id}}
	items, err := c.ListConfiguratorTemplates(ctx, q)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].ID.String() == id {
			return &items[i], nil
		}
	}
	return nil, &APIError{StatusCode: 404, Status: "404 Not Found", Method: "GET", Path: "devices/configurator/templates/" + id, Body: []byte("not found")}
}
