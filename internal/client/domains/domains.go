// Package domains wraps the mailbox.org domain.* API methods.
package domains

import (
	"context"
	"errors"
	"strings"

	"github.com/charpand/terraform-provider-mailboxorg/internal/client"
	"github.com/charpand/terraform-provider-mailboxorg/internal/client/jsonx"
)

// Domain is the result of domain.get.
type Domain struct {
	Domain         string         `json:"domain"`
	Status         jsonx.Text     `json:"status"`
	ContextID      jsonx.Text     `json:"context_id"`
	DateRegister   jsonx.Text     `json:"date_register"`
	Memo           jsonx.Text     `json:"memo"`
	PlansAvailable jsonx.TextList `json:"plansavailable"`
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

// Entry is one row of domain.list.
type Entry struct {
	Domain     string      `json:"domain"`
	CountMails jsonx.Count `json:"count_mails"`
}

// List returns the domains of an account (domain.list). filter may be empty.
func List(ctx context.Context, c *client.Client, account, filter string) ([]Entry, error) {
	if strings.TrimSpace(account) == "" {
		return nil, errors.New("domains: account must not be empty")
	}
	params := map[string]string{"account": account}
	if filter != "" {
		params["filter"] = filter
	}
	var entries []Entry
	if err := c.Read(ctx, "domain.list", params, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}
