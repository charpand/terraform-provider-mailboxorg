// Package domains wraps the mailbox.org domain.* API methods.
package domains

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/charpand/terraform-provider-mailboxorg/internal/client"
)

// Text is a JSON scalar rendered as text. The API documentation does not state
// the JSON types of domain.get fields, so strings, numbers and booleans are all
// accepted; null becomes the empty string.
type Text string

// UnmarshalJSON implements json.Unmarshaler.
func (t *Text) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if bytes.Equal(b, []byte("null")) {
		*t = ""
		return nil
	}
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*t = Text(s)
		return nil
	}
	*t = Text(b)
	return nil
}

// Domain is the result of domain.get.
type Domain struct {
	Domain         string `json:"domain"`
	Status         Text   `json:"status"`
	ContextID      Text   `json:"context_id"`
	DateRegister   Text   `json:"date_register"`
	Memo           Text   `json:"memo"`
	PlansAvailable []Text `json:"plansavailable"`
}

// Get returns the details of one domain. The API error for an unknown domain is
// returned unchanged, because its code is not documented.
func Get(ctx context.Context, c *client.Client, domain string) (*Domain, error) {
	if strings.TrimSpace(domain) == "" {
		return nil, errors.New("domains: domain must not be empty")
	}
	var d Domain
	if err := c.Read(ctx, "domain.get", map[string]string{"domain": domain}, &d); err != nil {
		return nil, err
	}
	return &d, nil
}
