// Package mails wraps the mailbox.org mail.* API methods.
package mails

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/charpand/terraform-provider-mailboxorg/internal/client"
	"github.com/charpand/terraform-provider-mailboxorg/internal/client/jsonx"
)

// Mail is the subset of mail.get that is exposed. The API documentation does
// not state field types, so tolerant jsonx types are used throughout. Postal
// address fields are intentionally not decoded.
type Mail struct {
	UID                  jsonx.Text     `json:"uid"`
	Mail                 string         `json:"mail"`
	Type                 jsonx.Text     `json:"type"`
	Plan                 jsonx.Text     `json:"plan"`
	PlansAvailable       jsonx.TextList `json:"plansavailable"`
	AdditionalMailQuota  jsonx.Text     `json:"additional_mail_quota"`
	AdditionalCloudQuota jsonx.Text     `json:"additional_cloud_quota"`
	FirstName            jsonx.Text     `json:"first_name"`
	LastName             jsonx.Text     `json:"last_name"`
	InboxSave            jsonx.Flag     `json:"inboxsave"`
	RequireResetPassword jsonx.Flag     `json:"require_reset_password"`
	Forwards             jsonx.TextList `json:"forwards"`
	Aliases              jsonx.TextList `json:"aliases"`
	AlternateMail        jsonx.Text     `json:"alternate_mail"`
	Memo                 jsonx.Text     `json:"memo"`
	AllowNets            jsonx.Text     `json:"allow_nets"`
	UIDExtern            jsonx.Text     `json:"uid_extern"`
	Active               jsonx.Flag     `json:"active"`
	Language             jsonx.Text     `json:"language"`
	Capabilities         jsonx.TextList `json:"capabilities"`
	CreationDateISO      jsonx.Text     `json:"creation_date_iso"`
	DeletionDate         jsonx.Text     `json:"deletion_date"`
	MailArchiveEnabled   jsonx.Flag     `json:"mailarchive_enabled"`
}

// Get returns one mail address (mail.get). The API error is returned unchanged:
// an address of another login gives code 8 (client.ErrForbidden), an unknown
// address gives code 6 with the detail "could not resolve mail_dn for <address>".
func Get(ctx context.Context, c *client.Client, mail string) (*Mail, error) {
	if strings.TrimSpace(mail) == "" {
		return nil, errors.New("mails: mail must not be empty")
	}
	var m Mail
	if err := c.Read(ctx, "mail.get", map[string]string{"mail": mail}, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Entry is one row of mail.list.
type Entry struct {
	Mail         string         `json:"mail"`
	Domain       jsonx.Text     `json:"domain"`
	Type         jsonx.Text     `json:"type"`
	Plan         jsonx.Text     `json:"plan"`
	Memo         jsonx.Text     `json:"memo"`
	Forwards     jsonx.TextList `json:"forwards"`
	Aliases      jsonx.TextList `json:"aliases"`
	Capabilities jsonx.TextList `json:"capabilities"`
	CreationDate jsonx.Text     `json:"creation_date"`
}

// ListOptions selects the addresses for List. At least one field is required.
type ListOptions struct {
	// Domain lists the addresses of one domain.
	Domain string
	// Query searches all domains; "%" is a wildcard.
	Query string
}

const (
	pageSize = 100
	// maxPages guards against a server that never reports the last page.
	maxPages = 1000
)

type page struct {
	TotalHits  jsonx.Count `json:"totalHits"`
	TotalPages jsonx.Count `json:"totalPages"`
	Results    []Entry     `json:"results"`
}

// List returns all addresses matching the options (mail.list), reading every
// page. Verified live: pages start at 1 (page 0 is error 462), a paginated
// answer is {totalHits, totalPages, results}, and a page past the end is a bare
// empty array.
func List(ctx context.Context, c *client.Client, opts ListOptions) ([]Entry, error) {
	if strings.TrimSpace(opts.Domain) == "" && strings.TrimSpace(opts.Query) == "" {
		return nil, errors.New("mails: domain or query is required")
	}
	var all []Entry
	for n := 1; n <= maxPages; n++ {
		params := map[string]any{"page_size": pageSize, "page": n}
		if opts.Domain != "" {
			params["domain"] = opts.Domain
		}
		if opts.Query != "" {
			params["query"] = opts.Query
		}
		var raw json.RawMessage
		if err := c.Read(ctx, "mail.list", params, &raw); err != nil {
			return nil, err
		}
		if bytes.HasPrefix(bytes.TrimSpace(raw), []byte("[")) {
			return all, nil // empty page: nothing (more) to read
		}
		var p page
		if err := json.Unmarshal(raw, &p); err != nil {
			return nil, fmt.Errorf("mailbox.org mail.list: unexpected result shape: %w: %v", client.ErrDecode, err)
		}
		all = append(all, p.Results...)
		if int64(p.TotalPages) <= int64(n) || len(p.Results) == 0 {
			return all, nil
		}
	}
	return nil, fmt.Errorf("mailbox.org mail.list: more than %d pages: %w", maxPages, client.ErrDecode)
}
