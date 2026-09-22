package rms

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

func (c *Client) ListDataCollectConfigs(ctx context.Context, q url.Values) ([]DataCollectConfig, error) {
	return ListAll[DataCollectConfig](ctx, c, "data-collect/configs", q)
}

func (c *Client) GetDataCollectConfig(ctx context.Context, id string) (*DataCollectConfig, error) {
	var env Envelope[DataCollectConfig]
	if err := c.Get(ctx, "data-collect/configs/"+id, nil, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (c *Client) CreateDataCollectConfig(ctx context.Context, payload json.RawMessage) (string, error) {
	var env struct {
		Data struct {
			ID any `json:"id"`
		} `json:"data"`
	}
	if err := c.Post(ctx, "data-collect/configs", payload, &env); err != nil {
		return "", err
	}
	if env.Data.ID == nil {
		return "", fmt.Errorf("rms: data-collect config create returned empty id")
	}
	return fmt.Sprint(env.Data.ID), nil
}

func (c *Client) UpdateDataCollectConfig(ctx context.Context, id string, payload json.RawMessage) error {
	return c.Put(ctx, "data-collect/configs/"+id, payload, nil)
}

func (c *Client) DeleteDataCollectConfig(ctx context.Context, id string) error {
	return c.Delete(ctx, "data-collect/configs/"+id, nil)
}
