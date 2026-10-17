package provider_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/charpand/terraform-provider-mailboxorg/internal/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const mailProviderConfig = `
provider "mailboxorg" {
  base_url = %q
  username = "user"
  password = "secret"
}

data "mailboxorg_mail" "this" {
  mail = "a@example.com"
}
`

func TestMailDataSource(t *testing.T) {
	srv := fakeAPI(t)
	srv.Handle("mail.get", func(testutils.Call) (any, *testutils.RPCError) {
		return map[string]any{
			"uid": 42, "mail": "a@example.com", "type": "inboxforward", "plan": "standard",
			"plansavailable": []string{"light", "standard"}, "first_name": "A", "last_name": "B",
			"inboxsave": "1", "require_reset_password": false, "forwards": []string{"x@example.org"},
			"aliases": []string{}, "memo": nil, "active": 1, "language": "de",
			"creation_date_iso": "2026-01-02", "mailarchive_enabled": true, "capabilities": []string{"MAIL_SPAMPROTECTION"},
			// Fields that must never be exposed.
			"street": "Secret Street 1", "birthday": "1990-01-01",
		}, nil
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6Factories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(mailProviderConfig, srv.URL),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.mailboxorg_mail.this", "id", "a@example.com"),
				resource.TestCheckResourceAttr("data.mailboxorg_mail.this", "uid", "42"),
				resource.TestCheckResourceAttr("data.mailboxorg_mail.this", "type", "inboxforward"),
				resource.TestCheckResourceAttr("data.mailboxorg_mail.this", "inboxsave", "true"),
				resource.TestCheckResourceAttr("data.mailboxorg_mail.this", "active", "true"),
				resource.TestCheckResourceAttr("data.mailboxorg_mail.this", "forwards.0", "x@example.org"),
				resource.TestCheckResourceAttr("data.mailboxorg_mail.this", "aliases.#", "0"),
				resource.TestCheckResourceAttr("data.mailboxorg_mail.this", "creation_date", "2026-01-02"),
				resource.TestCheckResourceAttr("data.mailboxorg_mail.this", "capabilities.0", "MAIL_SPAMPROTECTION"),
				resource.TestCheckNoResourceAttr("data.mailboxorg_mail.this", "street"),
			),
		}},
	})
}

func TestMailDataSourceExplainsNoPermission(t *testing.T) {
	srv := fakeAPI(t)
	srv.Handle("mail.get", func(testutils.Call) (any, *testutils.RPCError) {
		return nil, &testutils.RPCError{Code: 8, Message: "keine Berechtigung"}
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6Factories,
		Steps: []resource.TestStep{{
			Config:      fmt.Sprintf(mailProviderConfig, srv.URL),
			ExpectError: regexp.MustCompile(`(?s)Mail address not found or not\s+manageable`),
		}},
	})
}

func TestMailDataSourceExplainsUnknownAddress(t *testing.T) {
	srv := fakeAPI(t)
	srv.Handle("mail.get", func(testutils.Call) (any, *testutils.RPCError) {
		return nil, &testutils.RPCError{Code: 6, Message: "Es gab Probleme beim Datenbankzugriff", Data: "could not resolve mail_dn for a@example.com"}
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6Factories,
		Steps: []resource.TestStep{{
			Config:      fmt.Sprintf(mailProviderConfig, srv.URL),
			ExpectError: regexp.MustCompile(`(?s)could not resolve mail_dn\s+for\s+a@example.com`),
		}},
	})
}
