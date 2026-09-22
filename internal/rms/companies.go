package rms

import (
	"context"
	"fmt"
	"net/url"
)

func (c *Client) ListCompanies(ctx context.Context, q url.Values) ([]Company, error) {
	return ListAll[Company](ctx, c, "companies", q)
}

func (c *Client) GetCompany(ctx context.Context, id string) (*Company, error) {
	var env Envelope[Company]
	if err := c.Get(ctx, "companies/"+id, nil, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (c *Client) CreateCompany(ctx context.Context, in CompanyCreate) (*Company, error) {
	var env Envelope[Company]
	if err := c.Post(ctx, "companies", in, &env); err != nil {
		return nil, err
	}
	if env.Data.ID.String() == "" {
		return nil, fmt.Errorf("rms: company create returned empty id")
	}
	return &env.Data, nil
}

func (c *Client) UpdateCompany(ctx context.Context, id string, in CompanyUpdate) error {
	return c.Put(ctx, "companies/"+id, in, nil)
}

func (c *Client) DeleteCompany(ctx context.Context, id string) error {
	return c.Delete(ctx, "companies/"+id, nil)
}
