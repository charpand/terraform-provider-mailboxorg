// Package client is a Terraform-independent client for the mailbox.org JSON-RPC API.
package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// DefaultBaseURL is the mailbox.org API endpoint.
const DefaultBaseURL = "https://api.mailbox.org/v1/"

// DefaultTimeout bounds a single HTTP request.
const DefaultTimeout = 60 * time.Second

// Config configures a Client.
type Config struct {
	BaseURL    string
	Username   string
	Password   string
	Timeout    time.Duration
	HTTPClient *http.Client
}

// Client talks to the mailbox.org API. It logs in lazily and shares one
// session between concurrent callers. It is safe for concurrent use.
type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client

	mu         sync.Mutex
	session    string
	generation uint64
}

// New builds a Client. No network call is made until the first request.
func New(cfg Config) (*Client, error) {
	if cfg.Username == "" || cfg.Password == "" {
		return nil, errors.New("client: username and password are required")
	}
	c := &Client{
		baseURL:    cfg.BaseURL,
		username:   cfg.Username,
		password:   cfg.Password,
		httpClient: cfg.HTTPClient,
	}
	if c.baseURL == "" {
		c.baseURL = DefaultBaseURL
	}
	if c.httpClient == nil {
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = DefaultTimeout
		}
		c.httpClient = &http.Client{Timeout: timeout}
	}
	return c, nil
}

// Read performs a repeatable (side-effect free) call. If the session turns out
// to be invalid it logs in again once and resends the request.
func (c *Client) Read(ctx context.Context, method string, params, out any) error {
	return c.call(ctx, method, params, out, false)
}

// Mutate performs a call with side effects. It is never resent automatically:
// an invalid session is reported, and a failure after sending yields ErrAmbiguous.
func (c *Client) Mutate(ctx context.Context, method string, params, out any) error {
	return c.call(ctx, method, params, out, true)
}

func (c *Client) call(ctx context.Context, method string, params, out any, mutating bool) error {
	session, gen, err := c.ensureSession(ctx)
	if err != nil {
		return err
	}
	raw, err := c.roundTrip(ctx, method, params, session, mutating)
	var apiErr *APIError
	if !mutating && errors.As(err, &apiErr) && apiErr.Code == codeInvalidAuthID {
		session, err = c.refreshSession(ctx, gen)
		if err != nil {
			return err
		}
		raw, err = c.roundTrip(ctx, method, params, session, mutating)
	}
	if err != nil {
		return err
	}
	return decodeResult(method, raw, out)
}

func decodeResult(method string, raw json.RawMessage, out any) error {
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("mailbox.org %s: unexpected result shape: %w: %v", method, ErrDecode, err)
	}
	return nil
}
