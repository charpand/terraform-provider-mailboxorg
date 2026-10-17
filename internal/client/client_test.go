package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/charpand/terraform-provider-mailboxorg/internal/client"
	"github.com/charpand/terraform-provider-mailboxorg/internal/testutils"
)

const secretPassword = "canary-password-123"

func newClient(t *testing.T, url string) *client.Client {
	t.Helper()
	c, err := client.New(client.Config{BaseURL: url, Username: "user", Password: secretPassword})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func authHandler(counter *atomic.Int32) testutils.Handler {
	return func(testutils.Call) (any, *testutils.RPCError) {
		n := counter.Add(1)
		return map[string]string{"session": "session-" + string(rune('0'+n)), "level": "admin"}, nil
	}
}

func TestReadLogsInLazilyAndSendsSession(t *testing.T) {
	srv := testutils.NewFakeServer(t)
	var logins atomic.Int32
	srv.Handle("auth", authHandler(&logins))
	srv.Handle("hello.innerworld", func(testutils.Call) (any, *testutils.RPCError) { return "hi", nil })

	c := newClient(t, srv.URL)
	if len(srv.Calls()) != 0 {
		t.Fatal("New must not call the API")
	}
	var out string
	for range 2 {
		if err := c.Read(context.Background(), "hello.innerworld", nil, &out); err != nil {
			t.Fatal(err)
		}
	}
	calls := srv.Calls()
	if logins.Load() != 1 || len(calls) != 3 {
		t.Fatalf("want 1 login and 3 calls, got %d logins and %d calls", logins.Load(), len(calls))
	}
	if calls[1].Session != "session-1" || calls[2].Session != "session-1" {
		t.Fatalf("session header not sent: %+v", calls)
	}
	if string(calls[1].Params) != `""` {
		t.Fatalf("nil params should be sent as empty string, got %s", calls[1].Params)
	}
	var authParams map[string]string
	if err := json.Unmarshal(calls[0].Params, &authParams); err != nil || authParams["user"] != "user" || authParams["pass"] != secretPassword {
		t.Fatalf("unexpected auth params %s", calls[0].Params)
	}
}

func TestReadReauthenticatesOnceOnInvalidSession(t *testing.T) {
	srv := testutils.NewFakeServer(t)
	var logins atomic.Int32
	srv.Handle("auth", authHandler(&logins))
	srv.Handle("domain.get", func(call testutils.Call) (any, *testutils.RPCError) {
		if call.Session == "session-1" {
			return nil, &testutils.RPCError{Code: 253, Message: "This Auth-ID is invalid"}
		}
		return map[string]string{"domain": "example.com"}, nil
	})

	var out map[string]string
	if err := newClient(t, srv.URL).Read(context.Background(), "domain.get", map[string]string{"domain": "example.com"}, &out); err != nil {
		t.Fatal(err)
	}
	if out["domain"] != "example.com" || logins.Load() != 2 {
		t.Fatalf("out=%v logins=%d", out, logins.Load())
	}
}

func TestReadGivesUpAfterSecondInvalidSession(t *testing.T) {
	srv := testutils.NewFakeServer(t)
	var logins atomic.Int32
	srv.Handle("auth", authHandler(&logins))
	srv.Handle("domain.get", func(testutils.Call) (any, *testutils.RPCError) {
		return nil, &testutils.RPCError{Code: 253, Message: "This Auth-ID is invalid"}
	})

	err := newClient(t, srv.URL).Read(context.Background(), "domain.get", nil, nil)
	if !errors.Is(err, client.ErrAuthentication) {
		t.Fatalf("want ErrAuthentication, got %v", err)
	}
	if logins.Load() != 2 {
		t.Fatalf("want exactly 2 logins, got %d", logins.Load())
	}
}

func TestMutateIsNeverReplayedOnInvalidSession(t *testing.T) {
	srv := testutils.NewFakeServer(t)
	var logins atomic.Int32
	srv.Handle("auth", authHandler(&logins))
	srv.Handle("mail.del", func(testutils.Call) (any, *testutils.RPCError) {
		return nil, &testutils.RPCError{Code: 253, Message: "This Auth-ID is invalid"}
	})

	err := newClient(t, srv.URL).Mutate(context.Background(), "mail.del", nil, nil)
	if !errors.Is(err, client.ErrAuthentication) {
		t.Fatalf("want ErrAuthentication, got %v", err)
	}
	var n int
	for _, call := range srv.Calls() {
		if call.Method == "mail.del" {
			n++
		}
	}
	if n != 1 || logins.Load() != 1 {
		t.Fatalf("mail.del sent %d times, %d logins", n, logins.Load())
	}
}

func TestConcurrentReadersShareOneRefresh(t *testing.T) {
	srv := testutils.NewFakeServer(t)
	var logins atomic.Int32
	srv.Handle("auth", authHandler(&logins))
	srv.Handle("hello.innerworld", func(call testutils.Call) (any, *testutils.RPCError) {
		if call.Session == "session-1" {
			return nil, &testutils.RPCError{Code: 253, Message: "This Auth-ID is invalid"}
		}
		return "hi", nil
	})

	c := newClient(t, srv.URL)
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			var out string
			if err := c.Read(context.Background(), "hello.innerworld", nil, &out); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if logins.Load() != 2 {
		t.Fatalf("want 2 logins (initial + one coalesced refresh), got %d", logins.Load())
	}
}

func TestBadCredentialsAreReportedWithoutSecrets(t *testing.T) {
	srv := testutils.NewFakeServer(t)
	srv.Handle("auth", func(testutils.Call) (any, *testutils.RPCError) {
		return nil, &testutils.RPCError{Code: 2, Message: "Access denied"}
	})

	err := newClient(t, srv.URL).Read(context.Background(), "hello.innerworld", nil, nil)
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != 2 || apiErr.Method != "auth" {
		t.Fatalf("want APIError code 2 from auth, got %v", err)
	}
	if !errors.Is(err, client.ErrAuthentication) {
		t.Fatal("want ErrAuthentication")
	}
	if strings.Contains(err.Error(), secretPassword) {
		t.Fatalf("error leaks the password: %v", err)
	}
}

func TestInvalidBodyIsDecodeError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("<html>not json</html>"))
	}))
	t.Cleanup(ts.Close)

	err := newClient(t, ts.URL).Read(context.Background(), "hello.world", nil, nil)
	if !errors.Is(err, client.ErrDecode) {
		t.Fatalf("want ErrDecode, got %v", err)
	}
}

func TestMutateTransportFailureIsAmbiguous(t *testing.T) {
	srv := testutils.NewFakeServer(t)
	var logins atomic.Int32
	srv.Handle("auth", authHandler(&logins))
	c := newClient(t, srv.URL)
	if err := c.Read(context.Background(), "hello.innerworld", nil, nil); err == nil {
		t.Fatal("expected method-not-found from fake server")
	}
	srv.Close()

	err := c.Mutate(context.Background(), "mail.set", nil, nil)
	if !errors.Is(err, client.ErrAmbiguous) {
		t.Fatalf("want ErrAmbiguous, got %v", err)
	}
	err = c.Read(context.Background(), "mail.get", nil, nil)
	if !errors.Is(err, client.ErrTransport) || errors.Is(err, client.ErrAmbiguous) {
		t.Fatalf("want plain ErrTransport for reads, got %v", err)
	}
}

func TestCancelledContextIsPassedThrough(t *testing.T) {
	srv := testutils.NewFakeServer(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := newClient(t, srv.URL).Read(ctx, "hello.world", nil, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("want context.Canceled, got %v", err)
	}
}

func TestNewRequiresCredentials(t *testing.T) {
	if _, err := client.New(client.Config{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestNoPermissionCodeMatchesErrForbidden(t *testing.T) {
	srv := testutils.NewFakeServer(t)
	var logins atomic.Int32
	srv.Handle("auth", authHandler(&logins))
	srv.Handle("domain.get", func(testutils.Call) (any, *testutils.RPCError) {
		return nil, &testutils.RPCError{Code: 8, Message: "Sie haben keine Berechtigung, diese Domain zu administrieren"}
	})

	err := newClient(t, srv.URL).Read(context.Background(), "domain.get", nil, nil)
	if !errors.Is(err, client.ErrForbidden) || errors.Is(err, client.ErrAuthentication) {
		t.Fatalf("want ErrForbidden only, got %v", err)
	}
	if logins.Load() != 1 {
		t.Fatalf("code 8 must not trigger a re-login, got %d logins", logins.Load())
	}
}
