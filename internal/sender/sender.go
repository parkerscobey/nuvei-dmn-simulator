package sender

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	ContentTypeFormURLEncoded = "application/x-www-form-urlencoded"
	defaultMaxResponseBytes   = 1 << 20
)

type Client struct {
	HTTPClient       *http.Client
	MaxResponseBytes int64
}

type Request struct {
	TargetURL      string
	EncodedPayload string
}

type Result struct {
	StatusCode int
	Body       string
}

func NewClient(httpClient *http.Client) *Client {
	return &Client{HTTPClient: httpClient}
}

func (c *Client) Send(ctx context.Context, req Request) (Result, error) {
	if req.TargetURL == "" {
		return Result{}, fmt.Errorf("target URL is required")
	}
	if req.EncodedPayload == "" {
		return Result{}, fmt.Errorf("encoded payload is required")
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, req.TargetURL, strings.NewReader(req.EncodedPayload))
	if err != nil {
		return Result{}, fmt.Errorf("build sender request: %w", err)
	}
	httpReq.Header.Set("Content-Type", ContentTypeFormURLEncoded)
	httpReq.Header.Set("Accept", "*/*")

	resp, err := c.httpClient().Do(httpReq)
	if err != nil {
		return Result{}, fmt.Errorf("send DMN request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, c.maxResponseBytes()))
	if err != nil {
		return Result{}, fmt.Errorf("read DMN response: %w", err)
	}

	return Result{
		StatusCode: resp.StatusCode,
		Body:       string(body),
	}, nil
}

func (c *Client) httpClient() *http.Client {
	if c != nil && c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

func (c *Client) maxResponseBytes() int64 {
	if c != nil && c.MaxResponseBytes > 0 {
		return c.MaxResponseBytes
	}
	return defaultMaxResponseBytes
}
