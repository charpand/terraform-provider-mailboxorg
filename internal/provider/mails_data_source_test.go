package provider_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/charpand/terraform-provider-mailboxorg/internal/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMailsDataSource(t *testing.T) {
	srv := fakeAPI(t)
	srv.Handle("mail.list", func(testutils.Call) (any, *testutils.RPCError) {
		return map[string]any{
			"totalHits": 2, "totalPages": 1,
			"results": []map[string]any{
				{"mail": "a@example.com", "domain": "example.com", "type": "inbox", "plan": "light", "memo": "",
					"forwards": []string{}, "aliases": []string{"b@example.com"}, "capabilities": []string{"MAIL_SPAMPROTECTION"}, "creation_date": "01.02.2026"},
				{"mail": "f@example.com", "domain": "example.com", "type": "forward", "plan": "light",
					"forwards": []string{"x@example.org"}, "aliases": []string{}, "capabilities": []string{}},
			},
		}, nil
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6Factories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`
provider "mailboxorg" {
  base_url = %q
  username = "user"
  password = "secret"
}

data "mailboxorg_mails" "this" {
  domain = "example.com"
}
`, srv.URL),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.mailboxorg_mails.this", "mails.#", "2"),
				resource.TestCheckResourceAttr("data.mailboxorg_mails.this", "mails.0.mail", "a@example.com"),
				resource.TestCheckResourceAttr("data.mailboxorg_mails.this", "mails.0.aliases.0", "b@example.com"),
				resource.TestCheckResourceAttr("data.mailboxorg_mails.this", "mails.1.forwards.0", "x@example.org"),
				resource.TestCheckResourceAttr("data.mailboxorg_mails.this", "mails.1.type", "forward"),
			),
		}},
	})
}

func TestMailsDataSourceNeedsDomainOrQuery(t *testing.T) {
	srv := fakeAPI(t)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6Factories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`
provider "mailboxorg" {
  base_url = %q
  username = "user"
  password = "secret"
}

data "mailboxorg_mails" "this" {}
`, srv.URL),
			ExpectError: regexp.MustCompile(`Set at least one of`),
		}},
	})
}
