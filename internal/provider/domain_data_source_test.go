package provider_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/charpand/terraform-provider-mailboxorg/internal/provider"
	"github.com/charpand/terraform-provider-mailboxorg/internal/testutils"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var protoV6Factories = map[string]func() (tfprotov6.ProviderServer, error){
	"mailboxorg": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func fakeAPI(t *testing.T) *testutils.FakeServer {
	t.Helper()
	srv := testutils.NewFakeServer(t)
	srv.Handle("auth", func(testutils.Call) (any, *testutils.RPCError) {
		return map[string]string{"session": "s", "level": "reseller"}, nil
	})
	srv.Handle("domain.get", func(testutils.Call) (any, *testutils.RPCError) {
		return map[string]any{
			"domain": "example.com", "status": "active", "context_id": 7,
			"date_register": "2026-01-02", "memo": "test", "plansavailable": []string{"basic", "premium"},
		}, nil
	})
	return srv
}

func TestDomainDataSource(t *testing.T) {
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

data "mailboxorg_domain" "this" {
  domain = "example.com"
}
`, srv.URL),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.mailboxorg_domain.this", "id", "example.com"),
				resource.TestCheckResourceAttr("data.mailboxorg_domain.this", "status", "active"),
				resource.TestCheckResourceAttr("data.mailboxorg_domain.this", "context_id", "7"),
				resource.TestCheckResourceAttr("data.mailboxorg_domain.this", "memo", "test"),
				resource.TestCheckResourceAttr("data.mailboxorg_domain.this", "plans_available.#", "2"),
				resource.TestCheckResourceAttr("data.mailboxorg_domain.this", "plans_available.1", "premium"),
			),
		}},
	})
}

func TestProviderConfiguredFromEnvironment(t *testing.T) {
	srv := fakeAPI(t)
	t.Setenv("MAILBOXORG_BASE_URL", srv.URL)
	t.Setenv("MAILBOXORG_USERNAME", "env-user")
	t.Setenv("MAILBOXORG_PASSWORD", "env-secret")
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6Factories,
		Steps: []resource.TestStep{{
			Config: `data "mailboxorg_domain" "this" { domain = "example.com" }`,
			Check:  resource.TestCheckResourceAttr("data.mailboxorg_domain.this", "id", "example.com"),
		}},
	})
	for _, call := range srv.Calls() {
		if call.Method == "auth" && string(call.Params) != `{"pass":"env-secret","user":"env-user"}` {
			t.Fatalf("unexpected auth params %s", call.Params)
		}
	}
}

func TestProviderRequiresCredentials(t *testing.T) {
	t.Setenv("MAILBOXORG_USERNAME", "")
	t.Setenv("MAILBOXORG_PASSWORD", "")
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6Factories,
		Steps: []resource.TestStep{{
			Config:      `data "mailboxorg_domain" "this" { domain = "example.com" }`,
			ExpectError: regexp.MustCompile(`Missing username`),
		}},
	})
}

func TestDomainDataSourceSurfacesAPIError(t *testing.T) {
	srv := fakeAPI(t)
	srv.Handle("domain.get", func(testutils.Call) (any, *testutils.RPCError) {
		return nil, &testutils.RPCError{Code: 1234, Message: "boom"}
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
data "mailboxorg_domain" "this" { domain = "nope.example" }
`, srv.URL),
			ExpectError: regexp.MustCompile(`error 1234: boom`),
		}},
	})
}
