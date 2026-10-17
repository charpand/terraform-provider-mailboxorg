// Package testutils provides helpers for tests that need a mailbox.org API.
package testutils

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// Call is a recorded JSON-RPC request.
type Call struct {
	Method  string
	Params  json.RawMessage
	Session string
}

// Handler answers one JSON-RPC call. It returns the result value or, when
// rpcErr is non-nil, a JSON-RPC error object.
type Handler func(call Call) (result any, rpcErr *RPCError)

// RPCError is a JSON-RPC error object.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// FakeServer is an httptest JSON-RPC server that records every call.
type FakeServer struct {
	*httptest.Server

	mu       sync.Mutex
	calls    []Call
	handlers map[string]Handler
}

// NewFakeServer starts a server that is closed when the test ends.
func NewFakeServer(t *testing.T) *FakeServer {
	t.Helper()
	f := &FakeServer{handlers: map[string]Handler{}}
	f.Server = httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(f.Close)
	return f
}

// Handle registers the handler for a method, replacing any previous one.
func (f *FakeServer) Handle(method string, h Handler) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.handlers[method] = h
}

// Calls returns a copy of the recorded calls.
func (f *FakeServer) Calls() []Call {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]Call(nil), f.calls...)
}

func (f *FakeServer) serve(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var req struct {
		Method string          `json:"method"`
		Params json.RawMessage `json:"params"`
		ID     string          `json:"id"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	call := Call{Method: req.Method, Params: req.Params, Session: r.Header.Get("HPLS-AUTH")}

	f.mu.Lock()
	f.calls = append(f.calls, call)
	h := f.handlers[req.Method]
	f.mu.Unlock()

	resp := map[string]any{"jsonrpc": "2.0", "id": req.ID}
	if h == nil {
		resp["error"] = RPCError{Code: -32601, Message: "Method not found"}
	} else if result, rpcErr := h(call); rpcErr != nil {
		resp["error"] = rpcErr
	} else {
		resp["result"] = result
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
