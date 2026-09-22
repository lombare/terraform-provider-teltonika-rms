package rms

import (
	"context"
	"fmt"
	"net/url"
)

func (c *Client) ListTags(ctx context.Context, q url.Values) ([]Tag, error) {
	return ListAll[Tag](ctx, c, "tags", q)
}

func (c *Client) GetTag(ctx context.Context, id string) (*Tag, error) {
	var env Envelope[Tag]
	if err := c.Get(ctx, "tags/"+id, nil, &env); err != nil {
		return nil, err
	}
	return &env.Data, nil
}

func (c *Client) CreateTag(ctx context.Context, in TagCreate) (*Tag, error) {
	var env Envelope[Tag]
	if err := c.Post(ctx, "tags", in, &env); err != nil {
		return nil, err
	}
	if env.Data.ID.String() == "" {
		return nil, fmt.Errorf("rms: tag create returned empty id")
	}
	return &env.Data, nil
}

func (c *Client) UpdateTag(ctx context.Context, id string, in TagUpdate) error {
	return c.Put(ctx, "tags/"+id, in, nil)
}

func (c *Client) DeleteTag(ctx context.Context, id string) error {
	return c.Delete(ctx, "tags/"+id, nil)
}

// AssignDeviceTags maps devices to tags. Empty slices are no-ops on the RMS side.
func (c *Client) AssignDeviceTags(ctx context.Context, deviceIDs, tagIDs []int64) error {
	body := map[string]any{
		"device_id": deviceIDs,
		"tag_id":    tagIDs,
	}
	return c.Post(ctx, "devices/tags/assign", body, nil)
}

// OverwriteDeviceTags replaces the tag set on the given devices with tagIDs.
func (c *Client) OverwriteDeviceTags(ctx context.Context, deviceIDs, tagIDs []int64) error {
	body := map[string]any{
		"device_id": deviceIDs,
		"tag_id":    tagIDs,
	}
	return c.Post(ctx, "devices/tags/overwrite", body, nil)
}

func (c *Client) UnassignDeviceTags(ctx context.Context, deviceIDs, tagIDs []int64) error {
	body := map[string]any{
		"device_id": deviceIDs,
		"tag_id":    tagIDs,
	}
	return c.Post(ctx, "devices/tags/unassign", body, nil)
}
