// Package mails wraps the mailbox.org mail.* API methods.
package mails

import (
	"context"
	"errors"
	"strings"

	"github.com/charpand/terraform-provider-mailboxorg/internal/client"
	"github.com/charpand/terraform-provider-mailboxorg/internal/client/jsonx"
)

// Mail is the subset of mail.get that is exposed. The API documentation does
// not state field types, so tolerant jsonx types are used throughout. Postal
// address fields and capabilities are intentionally not decoded yet.
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
	CreationDateISO      jsonx.Text     `json:"creation_date_iso"`
	DeletionDate         jsonx.Text     `json:"deletion_date"`
	MailArchiveEnabled   jsonx.Flag     `json:"mailarchive_enabled"`
}

// Get returns one mail address (mail.get). The API error for an unknown or
// foreign address is returned unchanged (see client.ErrForbidden).
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
