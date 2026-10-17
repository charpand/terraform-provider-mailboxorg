package provider_test

import (
	"fmt"
	"testing"

	"github.com/charpand/terraform-provider-mailboxorg/internal/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDomainsDataSource(t *testing.T) {
	srv := fakeAPI(t)
	srv.Handle("domain.list", func(testutils.Call) (any, *testutils.RPCError) {
		return []map[string]any{{"domain": "a.example", "count_mails": 3}, {"domain": "b.example", "count_mails": "0"}}, nil
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

data "mailboxorg_domains" "this" {
  account = "acme"
  filter  = "example"
}
`, srv.URL),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.mailboxorg_domains.this", "id", "acme"),
				resource.TestCheckResourceAttr("data.mailboxorg_domains.this", "domains.#", "2"),
				resource.TestCheckResourceAttr("data.mailboxorg_domains.this", "domains.0.domain", "a.example"),
				resource.TestCheckResourceAttr("data.mailboxorg_domains.this", "domains.0.count_mails", "3"),
				resource.TestCheckResourceAttr("data.mailboxorg_domains.this", "domains.1.count_mails", "0"),
			),
		}},
	})
	var listCall testutils.Call
	for _, c := range srv.Calls() {
		if c.Method == "domain.list" {
			listCall = c
		}
	}
	if string(listCall.Params) != `{"account":"acme","filter":"example"}` {
		t.Fatalf("unexpected params %s", listCall.Params)
	}
}
