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

func TestGetDecodesRealResponseShape(t *testing.T) {
	c := setup(t, `{"domain":"example.com","context_id":"306297","status":"registriert","date_register":"2026-02-01","memo":"","plansavailable":["light","standard","premium"]}`)
	d, err := domains.Get(context.Background(), c, "example.com")
	if err != nil {
		t.Fatal(err)
	}
	if d.Status != "registriert" || d.ContextID != "306297" || d.DateRegister != "2026-02-01" || len(d.PlansAvailable) != 3 {
		t.Fatalf("unexpected domain %+v", d)
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

func TestListSendsFilterAndDecodes(t *testing.T) {
	srv := testutils.NewFakeServer(t)
	srv.Handle("auth", func(testutils.Call) (any, *testutils.RPCError) {
		return map[string]string{"session": "s", "level": "account"}, nil
	})
	srv.Handle("domain.list", func(testutils.Call) (any, *testutils.RPCError) {
		return json.RawMessage(`[{"domain":"a.example","count_mails":3},{"domain":"b.example","count_mails":"0"}]`), nil
	})
	c, err := client.New(client.Config{BaseURL: srv.URL, Username: "u", Password: "p"})
	if err != nil {
		t.Fatal(err)
	}

	got, err := domains.List(context.Background(), c, "acme", "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Domain != "a.example" || got[0].CountMails != 3 || got[1].CountMails != 0 {
		t.Fatalf("unexpected entries %+v", got)
	}
	calls := srv.Calls()
	if string(calls[len(calls)-1].Params) != `{"account":"acme","filter":"a"}` {
		t.Fatalf("unexpected params %s", calls[len(calls)-1].Params)
	}
	if _, err := domains.List(context.Background(), c, " ", ""); err == nil {
		t.Fatal("expected error for empty account")
	}
}
