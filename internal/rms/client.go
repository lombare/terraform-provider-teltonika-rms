package rms

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultBaseURL = "https://rms.teltonika-networks.com/api"
	UserAgent      = "terraform-provider-teltonika-rms"
)

type Client struct {
	baseURL    *url.URL
	token      string
	httpClient *http.Client
}

func NewClient(baseURL, token string) (*Client, error) {
	if token == "" {
		return nil, errors.New("rms: token is required")
	}
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("rms: invalid base URL %q: %w", baseURL, err)
	}
	return &Client{
		baseURL:    u,
		token:      token,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}, nil
}

// APIError is returned when the RMS API responds with a non-2xx status.
type APIError struct {
	StatusCode int
	Status     string
	Path       string
	Method     string
	Body       []byte
}

func (e *APIError) Error() string {
	body := strings.TrimSpace(string(e.Body))
	if len(body) > 512 {
		body = body[:512] + "…"
	}
	return fmt.Sprintf("rms: %s %s: %s: %s", e.Method, e.Path, e.Status, body)
}

func (e *APIError) NotFound() bool { return e != nil && e.StatusCode == http.StatusNotFound }

func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.NotFound()
	}
	return false
}

// do performs an HTTP request and decodes the JSON response body into out (if non-nil).
// query is optional; body is optional.
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	rel, err := url.Parse(path)
	if err != nil {
		return fmt.Errorf("rms: invalid path %q: %w", path, err)
	}
	u := c.baseURL.ResolveReference(rel)
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}

	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("rms: encode body: %w", err)
		}
		reader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reader)
	if err != nil {
		return fmt.Errorf("rms: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", UserAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("rms: %s %s: %w", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("rms: read body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Path:       path,
			Method:     method,
			Body:       raw,
		}
	}

	if out == nil || len(raw) == 0 || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("rms: decode response for %s %s: %w", method, path, err)
	}
	return nil
}

// Get performs a GET request; decodes into out if non-nil.
func (c *Client) Get(ctx context.Context, path string, query url.Values, out any) error {
	return c.do(ctx, http.MethodGet, path, query, nil, out)
}

// Post performs a POST request with a JSON body; decodes into out if non-nil.
func (c *Client) Post(ctx context.Context, path string, body any, out any) error {
	return c.do(ctx, http.MethodPost, path, nil, body, out)
}

// Put performs a PUT request with a JSON body; decodes into out if non-nil.
func (c *Client) Put(ctx context.Context, path string, body any, out any) error {
	return c.do(ctx, http.MethodPut, path, nil, body, out)
}

// Patch performs a PATCH request with a JSON body; decodes into out if non-nil.
func (c *Client) Patch(ctx context.Context, path string, body any, out any) error {
	return c.do(ctx, http.MethodPatch, path, nil, body, out)
}

// Delete performs a DELETE request; decodes into out if non-nil.
func (c *Client) Delete(ctx context.Context, path string, out any) error {
	return c.do(ctx, http.MethodDelete, path, nil, nil, out)
}

// ListPage is the RMS list-response envelope: `{ data, meta: { total, ... } }`.
type ListPage[T any] struct {
	Data []T  `json:"data"`
	Meta Meta `json:"meta"`
}

type Meta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// Envelope is the RMS single-item response envelope: `{ data }`.
type Envelope[T any] struct {
	Data T `json:"data"`
}

// ListAll transparently pages through a list endpoint and returns every item.
// The endpoint must accept `limit` and `offset` query params (default across RMS).
func ListAll[T any](ctx context.Context, c *Client, path string, extra url.Values) ([]T, error) {
	const pageSize = 100
	offset := 0
	all := make([]T, 0, pageSize)
	for {
		q := url.Values{}
		for k, vs := range extra {
			for _, v := range vs {
				q.Add(k, v)
			}
		}
		q.Set("limit", strconv.Itoa(pageSize))
		q.Set("offset", strconv.Itoa(offset))
		var page ListPage[T]
		if err := c.Get(ctx, path, q, &page); err != nil {
			return nil, err
		}
		all = append(all, page.Data...)
		if len(page.Data) < pageSize {
			return all, nil
		}
		offset += len(page.Data)
	}
}
