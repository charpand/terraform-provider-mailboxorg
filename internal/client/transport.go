package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync/atomic"
)

const sessionHeader = "HPLS-AUTH"

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
	ID      string `json:"id"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *rpcError       `json:"error"`
}

var requestID atomic.Uint64

// roundTrip performs one JSON-RPC call. A nil params is sent as "" as in the
// documented examples for methods without parameters. mutating decides how a
// failure after the request was sent is classified.
func (c *Client) roundTrip(ctx context.Context, method string, params any, session string, mutating bool) (json.RawMessage, error) {
	if params == nil {
		params = ""
	}
	body, err := json.Marshal(rpcRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      strconv.FormatUint(requestID.Add(1), 10),
	})
	if err != nil {
		return nil, fmt.Errorf("mailbox.org %s: encode request: %w", method, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("mailbox.org %s: build request: %w", method, err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if session != "" {
		req.Header.Set(sessionHeader, session)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil && !mutating {
			return nil, ctxErr
		}
		return nil, transportErr(method, err, mutating)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, transportErr(method, err, mutating)
	}

	var out rpcResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("mailbox.org %s: http status %d: %w", method, resp.StatusCode, ErrTransport)
		}
		return nil, fmt.Errorf("mailbox.org %s: invalid response: %w: %v", method, ErrDecode, err)
	}
	if out.Error != nil {
		return nil, &APIError{Method: method, Code: out.Error.Code, Message: out.Error.Message, Data: out.Error.Data}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mailbox.org %s: http status %d: %w", method, resp.StatusCode, ErrTransport)
	}
	if out.Result == nil {
		return nil, fmt.Errorf("mailbox.org %s: response has neither result nor error: %w", method, ErrDecode)
	}
	return out.Result, nil
}

func transportErr(method string, err error, mutating bool) error {
	if mutating {
		return fmt.Errorf("mailbox.org %s: %w: %w", method, ErrAmbiguous, err)
	}
	return fmt.Errorf("mailbox.org %s: %w: %w", method, ErrTransport, err)
}
