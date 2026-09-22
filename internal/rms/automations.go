package rms

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

func (c *Client) ListAutomations(ctx context.Context, q url.Values) ([]Automation, error) {
	return ListAll[Automation](ctx, c, "automations", q)
}

func (c *Client) GetAutomation(ctx context.Context, id string) (*Automation, error) {
	var env Envelope[Automation]
	if err := c.Get(ctx, "automations/"+id, nil, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (c *Client) CreateAutomation(ctx context.Context, payload json.RawMessage) (string, error) {
	var env struct {
		Data struct {
			ID any `json:"id"`
		} `json:"data"`
	}
	if err := c.Post(ctx, "automations", payload, &env); err != nil {
		return "", err
	}
	if env.Data.ID == nil {
		return "", fmt.Errorf("rms: automation create returned empty id")
	}
	return fmt.Sprint(env.Data.ID), nil
}

func (c *Client) UpdateAutomation(ctx context.Context, id string, payload json.RawMessage) error {
	return c.Put(ctx, "automations/"+id, payload, nil)
}

func (c *Client) DeleteAutomation(ctx context.Context, id string) error {
	return c.Delete(ctx, "automations/"+id, nil)
}
