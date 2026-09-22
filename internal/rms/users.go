package rms

import (
	"context"
	"fmt"
	"net/url"
)

func (c *Client) CurrentUser(ctx context.Context) (*User, error) {
	var env Envelope[User]
	if err := c.Get(ctx, "user", nil, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (c *Client) ListUsers(ctx context.Context, q url.Values) ([]User, error) {
	return ListAll[User](ctx, c, "users", q)
}

func (c *Client) GetUser(ctx context.Context, id string) (*User, error) {
	var env Envelope[User]
	if err := c.Get(ctx, "users/"+id, nil, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (c *Client) UpdateUser(ctx context.Context, id string, in UserUpdate) error {
	return c.Put(ctx, "users/"+id, in, nil)
}

// InviteUser posts to /users/invite. Returns the invitation id.
func (c *Client) InviteUser(ctx context.Context, in InviteUser) (string, error) {
	var env struct {
		Data struct {
			ID any `json:"id"`
		} `json:"data"`
	}
	if err := c.Post(ctx, "users/invite", in, &env); err != nil {
		return "", err
	}
	if env.Data.ID == nil {
		return "", fmt.Errorf("rms: invitation returned empty id")
	}
	return fmt.Sprint(env.Data.ID), nil
}

func (c *Client) DeleteInvitation(ctx context.Context, id string) error {
	return c.Delete(ctx, "users/invitations/"+id, nil)
}
