package rms

import (
	"context"
	"fmt"
	"net/url"
)

func (c *Client) ListRoles(ctx context.Context, q url.Values) ([]Role, error) {
	return ListAll[Role](ctx, c, "roles", q)
}

func (c *Client) GetRole(ctx context.Context, id string) (*Role, error) {
	var env Envelope[Role]
	if err := c.Get(ctx, "roles/"+id, nil, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (c *Client) CreateRole(ctx context.Context, in RoleCreate) (*Role, error) {
	var env Envelope[Role]
	if err := c.Post(ctx, "roles", in, &env); err != nil {
		return nil, err
	}
	if env.Data.ID.String() == "" {
		return nil, fmt.Errorf("rms: role create returned empty id")
	}
	return &env.Data, nil
}

func (c *Client) UpdateRole(ctx context.Context, in RoleUpdate) error {
	return c.Put(ctx, fmt.Sprintf("roles/%d", in.ID), in, nil)
}

func (c *Client) DeleteRole(ctx context.Context, id string) error {
	return c.Delete(ctx, "roles/"+id, nil)
}

func (c *Client) ListPermissions(ctx context.Context) ([]Permission, error) {
	return ListAll[Permission](ctx, c, "roles/permissions", nil)
}

func (c *Client) RolePermissions(ctx context.Context, id string) ([]Permission, error) {
	var env struct {
		Data []Permission `json:"data"`
	}
	if err := c.Get(ctx, "roles/"+id+"/permissions", nil, &env); err != nil {
		return nil, err
	}
	return env.Data, nil
}
