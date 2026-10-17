package mails_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/charpand/terraform-provider-mailboxorg/internal/client"
	"github.com/charpand/terraform-provider-mailboxorg/internal/client/mails"
	"github.com/charpand/terraform-provider-mailboxorg/internal/testutils"
)

func setup(t *testing.T, result string, rpcErr *testutils.RPCError) *client.Client {
	t.Helper()
	srv := testutils.NewFakeServer(t)
	srv.Handle("auth", func(testutils.Call) (any, *testutils.RPCError) {
		return map[string]string{"session": "s", "level": "account"}, nil
	})
	srv.Handle("mail.get", func(testutils.Call) (any, *testutils.RPCError) {
		if rpcErr != nil {
			return nil, rpcErr
		}
		return json.RawMessage(result), nil
	})
	c, err := client.New(client.Config{BaseURL: srv.URL, Username: "u", Password: "p"})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestGetDecodesLooselyTypedFields(t *testing.T) {
	c := setup(t, `{"uid":42,"mail":"a@example.com","type":"inboxforward","plan":"standard","plansavailable":["light","standard"],
		"first_name":"A","last_name":"B","inboxsave":"1","require_reset_password":false,"forwards":["x@example.org"],
		"aliases":[],"memo":null,"active":1,"language":"de","creation_date_iso":"2026-01-02","mailarchive_enabled":true}`, nil)
	m, err := mails.Get(context.Background(), c, "a@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if m.UID != "42" || m.Type != "inboxforward" || !m.InboxSave || m.RequireResetPassword || !m.Active || !m.MailArchiveEnabled {
		t.Fatalf("unexpected mail %+v", m)
	}
	if got := m.Forwards.Strings(); len(got) != 1 || got[0] != "x@example.org" || len(m.Aliases) != 0 || m.Memo != "" {
		t.Fatalf("unexpected lists %+v", m)
	}
}

func TestGetSurfacesForbidden(t *testing.T) {
	c := setup(t, "", &testutils.RPCError{Code: 8, Message: "keine Berechtigung"})
	if _, err := mails.Get(context.Background(), c, "x@example.com"); !errors.Is(err, client.ErrForbidden) {
		t.Fatalf("want ErrForbidden, got %v", err)
	}
}

func TestGetRejectsEmpty(t *testing.T) {
	c := setup(t, `{}`, nil)
	if _, err := mails.Get(context.Background(), c, ""); err == nil {
		t.Fatal("expected error")
	}
}

func TestListReadsAllPages(t *testing.T) {
	srv := testutils.NewFakeServer(t)
	srv.Handle("auth", func(testutils.Call) (any, *testutils.RPCError) {
		return map[string]string{"session": "s", "level": "account"}, nil
	})
	srv.Handle("mail.list", func(call testutils.Call) (any, *testutils.RPCError) {
		var p map[string]any
		_ = json.Unmarshal(call.Params, &p)
		switch p["page"] {
		case float64(1):
			return json.RawMessage(`{"totalHits":3,"totalPages":2,"results":[{"mail":"a@example.com","type":"inbox","forwards":["x@example.org"]},{"mail":"b@example.com"}]}`), nil
		case float64(2):
			return json.RawMessage(`{"totalHits":3,"totalPages":2,"results":[{"mail":"c@example.com","aliases":["d@example.com"]}]}`), nil
		}
		return json.RawMessage(`[]`), nil
	})
	c, err := client.New(client.Config{BaseURL: srv.URL, Username: "u", Password: "p"})
	if err != nil {
		t.Fatal(err)
	}

	got, err := mails.List(context.Background(), c, mails.ListOptions{Domain: "example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0].Mail != "a@example.com" || got[2].Mail != "c@example.com" || got[0].Forwards.Strings()[0] != "x@example.org" {
		t.Fatalf("unexpected entries %+v", got)
	}
	var pages int
	for _, call := range srv.Calls() {
		if call.Method == "mail.list" {
			pages++
		}
	}
	if pages != 2 {
		t.Fatalf("want 2 page requests, got %d", pages)
	}
}

func TestListHandlesBareEmptyArray(t *testing.T) {
	srv := testutils.NewFakeServer(t)
	srv.Handle("auth", func(testutils.Call) (any, *testutils.RPCError) {
		return map[string]string{"session": "s", "level": "account"}, nil
	})
	srv.Handle("mail.list", func(testutils.Call) (any, *testutils.RPCError) { return json.RawMessage(`[]`), nil })
	c, err := client.New(client.Config{BaseURL: srv.URL, Username: "u", Password: "p"})
	if err != nil {
		t.Fatal(err)
	}
	got, err := mails.List(context.Background(), c, mails.ListOptions{Query: "%"})
	if err != nil || len(got) != 0 {
		t.Fatalf("got %v, %v", got, err)
	}
	if _, err := mails.List(context.Background(), c, mails.ListOptions{}); err == nil {
		t.Fatal("expected error without domain or query")
	}
}
