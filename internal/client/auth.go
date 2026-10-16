package client

import (
	"context"
	"fmt"
)

type authResult struct {
	Session string `json:"session"`
	Level   string `json:"level"`
}

// ensureSession returns the current session, logging in if there is none.
func (c *Client) ensureSession(ctx context.Context) (string, uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.session == "" {
		if err := c.loginLocked(ctx); err != nil {
			return "", 0, err
		}
	}
	return c.session, c.generation, nil
}

// refreshSession replaces a session that failed with seenGeneration. If another
// caller already replaced it, that newer session is reused instead of logging in again.
func (c *Client) refreshSession(ctx context.Context, seenGeneration uint64) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.generation != seenGeneration && c.session != "" {
		return c.session, nil
	}
	if err := c.loginLocked(ctx); err != nil {
		return "", err
	}
	return c.session, nil
}

// loginLocked calls auth; the caller must hold c.mu. The credentials never
// appear in returned errors, which only carry the method name and API message.
func (c *Client) loginLocked(ctx context.Context) error {
	c.session = ""
	raw, err := c.roundTrip(ctx, "auth", map[string]string{"user": c.username, "pass": c.password}, "", false)
	if err != nil {
		return err
	}
	var res authResult
	if err := decodeResult("auth", raw, &res); err != nil {
		return err
	}
	if res.Session == "" {
		return fmt.Errorf("mailbox.org auth: empty session in response: %w", ErrDecode)
	}
	c.session = res.Session
	c.generation++
	return nil
}
