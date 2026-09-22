package rms

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

func (c *Client) ListAlerts(ctx context.Context, q url.Values) ([]Alert, error) {
	return ListAll[Alert](ctx, c, "alerts", q)
}

func (c *Client) GetAlert(ctx context.Context, id string) (*Alert, error) {
	var env Envelope[Alert]
	if err := c.Get(ctx, "alerts/"+id, nil, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (c *Client) ListAlertConfigurations(ctx context.Context, q url.Values) ([]AlertConfiguration, error) {
	return ListAll[AlertConfiguration](ctx, c, "alerts-configurations", q)
}

func (c *Client) GetAlertConfiguration(ctx context.Context, id string) (*AlertConfiguration, error) {
	var env Envelope[AlertConfiguration]
	if err := c.Get(ctx, "alerts-configurations/"+id, nil, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

// CreateAlertConfiguration wraps the body as `{data: [payload]}` per the OpenAPI shape.
func (c *Client) CreateAlertConfiguration(ctx context.Context, payload json.RawMessage) (string, error) {
	body := map[string]any{"data": []json.RawMessage{payload}}
	var env struct {
		Data []struct {
			ID any `json:"id"`
		} `json:"data"`
	}
	if err := c.Post(ctx, "alerts-configurations", body, &env); err != nil {
		return "", err
	}
	if len(env.Data) == 0 || env.Data[0].ID == nil {
		return "", fmt.Errorf("rms: alert configuration create returned empty id")
	}
	return fmt.Sprint(env.Data[0].ID), nil
}

func (c *Client) UpdateAlertConfiguration(ctx context.Context, id string, payload json.RawMessage) error {
	return c.Put(ctx, "alerts-configurations/"+id, payload, nil)
}

func (c *Client) DeleteAlertConfiguration(ctx context.Context, id string) error {
	return c.Delete(ctx, "alerts-configurations/"+id, nil)
}
