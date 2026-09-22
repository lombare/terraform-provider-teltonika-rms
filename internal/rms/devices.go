package rms

import (
	"context"
	"encoding/json"
	"net/url"
)

func (c *Client) ListDevices(ctx context.Context, q url.Values) ([]Device, error) {
	return ListAll[Device](ctx, c, "devices", q)
}

func (c *Client) GetDevice(ctx context.Context, id string) (*Device, error) {
	var env Envelope[Device]
	if err := c.Get(ctx, "devices/"+id, nil, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (c *Client) UpdateDevice(ctx context.Context, id string, in DeviceUpdate) error {
	return c.Put(ctx, "devices/"+id, in, nil)
}

// ListDeviceMonitoring surfaces the /devices/monitoring endpoint (paginated).
func (c *Client) ListDeviceMonitoring(ctx context.Context, q url.Values) ([]DeviceMonitoring, error) {
	return ListAll[DeviceMonitoring](ctx, c, "devices/monitoring", q)
}

// DeviceStatistics returns the raw payload from /devices/statistics.
func (c *Client) DeviceStatistics(ctx context.Context, q url.Values) (json.RawMessage, error) {
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := c.Get(ctx, "devices/statistics", q, &env); err != nil {
		return nil, err
	}
	return env.Data, nil
}

// OnlineDeviceStatistics returns the raw payload from /statistics/devices/online.
func (c *Client) OnlineDeviceStatistics(ctx context.Context, q url.Values) (json.RawMessage, error) {
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := c.Get(ctx, "statistics/devices/online", q, &env); err != nil {
		return nil, err
	}
	return env.Data, nil
}
