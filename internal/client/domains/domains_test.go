package domains_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/charpand/terraform-provider-mailboxorg/internal/client"
	"github.com/charpand/terraform-provider-mailboxorg/internal/client/domains"
	"github.com/charpand/terraform-provider-mailboxorg/internal/testutils"
)

func setup(t *testing.T, result string) *client.Client {
	t.Helper()
	srv := testutils.NewFakeServer(t)
	srv.Handle("auth", func(testutils.Call) (any, *testutils.RPCError) {
		return map[string]string{"session": "s", "level": "reseller"}, nil
	})
	srv.Handle("domain.get", func(call testutils.Call) (any, *testutils.RPCError) {
		var p map[string]string
		_ = json.Unmarshal(call.Params, &p)
		if p["domain"] != "example.com" {
			return nil, &testutils.RPCError{Code: 404, Message: "unknown domain"}
		}
		return json.RawMessage(result), nil
	})
	c, err := client.New(client.Config{BaseURL: srv.URL, Username: "u", Password: "p"})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestGetDecodesStringsNumbersAndNull(t *testing.T) {
	c := setup(t, `{"domain":"example.com","status":1,"context_id":"42","date_register":null,"memo":"hi","plansavailable":["basic",2]}`)
	d, err := domains.Get(context.Background(), c, "example.com")
	if err != nil {
		t.Fatal(err)
	}
	if d.Domain != "example.com" || d.Status != "1" || d.ContextID != "42" || d.DateRegister != "" || d.Memo != "hi" {
		t.Fatalf("unexpected domain %+v", d)
	}
	if len(d.PlansAvailable) != 2 || d.PlansAvailable[0] != "basic" || d.PlansAvailable[1] != "2" {
		t.Fatalf("unexpected plans %+v", d.PlansAvailable)
	}
}

func TestGetSurfacesAPIError(t *testing.T) {
	c := setup(t, `{}`)
	_, err := domains.Get(context.Background(), c, "missing.example")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetRejectsEmptyDomain(t *testing.T) {
	c := setup(t, `{}`)
	if _, err := domains.Get(context.Background(), c, " "); err == nil {
		t.Fatal("expected error")
	}
}
