package rms

import (
	"context"
	"fmt"
	"net/url"
)

func (c *Client) ListEmailConfigurations(ctx context.Context, q url.Values) ([]EmailConfiguration, error) {
	return ListAll[EmailConfiguration](ctx, c, "email-configurations", q)
}

func (c *Client) GetEmailConfiguration(ctx context.Context, id string) (*EmailConfiguration, error) {
	var env Envelope[EmailConfiguration]
	if err := c.Get(ctx, "email-configurations/"+id, nil, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (c *Client) CreateEmailConfiguration(ctx context.Context, in EmailConfigurationCreate) (*EmailConfiguration, error) {
	var env Envelope[EmailConfiguration]
	if err := c.Post(ctx, "email-configurations", in, &env); err != nil {
		return nil, err
	}
	if env.Data.ID.String() == "" {
		return nil, fmt.Errorf("rms: email configuration create returned empty id")
	}
	return &env.Data, nil
}

func (c *Client) UpdateEmailConfiguration(ctx context.Context, id string, in EmailConfigurationUpdate) error {
	return c.Put(ctx, "email-configurations/"+id, in, nil)
}

func (c *Client) DeleteEmailConfiguration(ctx context.Context, id string) error {
	return c.Delete(ctx, "email-configurations/"+id, nil)
}
