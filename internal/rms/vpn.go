package rms

import (
	"context"
	"fmt"
	"net/url"
)

func (c *Client) ListVPNHubs(ctx context.Context, q url.Values) ([]VPNHub, error) {
	return ListAll[VPNHub](ctx, c, "vpn/hubs", q)
}

func (c *Client) GetVPNHub(ctx context.Context, id string) (*VPNHub, error) {
	// The RMS API exposes the hub metadata via /vpn/hubs/{id}/info.
	var env Envelope[VPNHub]
	if err := c.Get(ctx, "vpn/hubs/"+id+"/info", nil, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (c *Client) CreateVPNHub(ctx context.Context, in VPNHubCreate) (*VPNHub, error) {
	var env Envelope[VPNHub]
	if err := c.Post(ctx, "vpn/hubs", in, &env); err != nil {
		return nil, err
	}
	if env.Data.ID.String() == "" {
		return nil, fmt.Errorf("rms: vpn hub create returned empty id")
	}
	return &env.Data, nil
}

func (c *Client) UpdateVPNHub(ctx context.Context, id string, in VPNHubUpdate) error {
	return c.Put(ctx, "vpn/hubs/"+id, in, nil)
}

func (c *Client) DeleteVPNHub(ctx context.Context, id string) error {
	// RMS uses a bulk delete envelope: {"id": [ids]}.
	body := map[string]any{"id": []string{id}}
	return c.do(ctx, "DELETE", "vpn/hubs", nil, body, nil)
}

func (c *Client) ToggleVPNHub(ctx context.Context, id string, enabled bool) error {
	body := map[string]any{"enabled": enabled}
	return c.Put(ctx, "vpn/hubs/"+id+"/toggle", body, nil)
}

func (c *Client) ListVPNHubUsers(ctx context.Context, q url.Values) ([]VPNHubUser, error) {
	return ListAll[VPNHubUser](ctx, c, "vpn/hubs/users", q)
}

func (c *Client) GetVPNHubUser(ctx context.Context, id string) (*VPNHubUser, error) {
	var env Envelope[VPNHubUser]
	if err := c.Get(ctx, "vpn/hubs/users/"+id, nil, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (c *Client) CreateVPNHubUser(ctx context.Context, in VPNHubUserCreate) (*VPNHubUser, error) {
	var env Envelope[VPNHubUser]
	if err := c.Post(ctx, "vpn/hubs/users", in, &env); err != nil {
		return nil, err
	}
	if env.Data.ID.String() == "" {
		return nil, fmt.Errorf("rms: vpn hub user create returned empty id")
	}
	return &env.Data, nil
}

func (c *Client) DeleteVPNHubUser(ctx context.Context, id string) error {
	return c.Delete(ctx, "vpn/hubs/users/"+id, nil)
}

func (c *Client) ToggleVPNHubUser(ctx context.Context, id string, enabled bool) error {
	body := map[string]any{"enabled": enabled}
	return c.Put(ctx, "vpn/hubs/users/"+id+"/toggle", body, nil)
}
