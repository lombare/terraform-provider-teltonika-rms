package rms

import (
	"context"
	"net/url"
)

func (c *Client) ListFiles(ctx context.Context, q url.Values) ([]File, error) {
	return ListAll[File](ctx, c, "files", q)
}

func (c *Client) GetFile(ctx context.Context, id string) (*File, error) {
	var env Envelope[File]
	if err := c.Get(ctx, "files/"+id, nil, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (c *Client) DeleteFile(ctx context.Context, id string) error {
	return c.Delete(ctx, "files/"+id, nil)
}
