package main

import (
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"
)

var handshakeTest = []struct {
	req    *http.Request
	status int
	// body string TODO(ben): Some errors have body messages. We should check those.
}{
	{defaultRequest(), 101},
	{withUnsupportedVersion(), 426},
	{withIncorrectUpgrade(), 400},
	{withIncorrectConnection(), 400},
	{withNoSecKey(), 400},
	{withIncorrectMethod(), 400},
	{withNoHost(), 400},
	{withNoRequestURI(), 400},

	// {withSecKeyTooLong(), 400},
}

func TestWebSocketHandShake(t *testing.T) {
	for _, tt := range handshakeTest {
		resp := getRequest(tt.req)
		if resp.StatusCode != tt.status {
			t.Errorf("got %d, want %d", resp.StatusCode, tt.status)
		}
	}
}

func TestHandshakeWithUnsupportedVersion(t *testing.T) {
	r := withUnsupportedVersion()
	resp := getRequest(r)
	v := resp.Header.Get("Sec-Websocket-Version")
	if v != "13" {
		t.Errorf("expected \"Sec-Websocket-Version\" header in response to have %q, got %q", "13", v)
	}
}

func TestAcknowledgedHandShakeResponse(t *testing.T) {
	r := defaultRequest()
	resp := getRequest(r)
	conn := resp.Header.Get("Connection")
	upg := resp.Header.Get("Upgrade")
	acc := resp.Header.Get("Sec-WebSocket-Accept")

	if conn != "Upgrade" || upg != "websocket" || acc != "s3pPLMBiTxaQ9kYGzzhZRbK+xOo=" {
		dump, _ := httputil.DumpResponse(resp, true)
		t.Errorf("incorrect handshake response from server. Received:\n%s", dump)
	}
}

// base case that is a good ws handshake request
func defaultRequest() *http.Request {
	r := httptest.NewRequest("GET", "http://localhost", nil)
	r.Header.Set("Upgrade", "websocket")
	r.Header.Set("Connection", "Upgrade")
	r.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")
	r.Header.Set("Sec-WebSocket-Version", "13")
	return r
}

func withUnsupportedVersion() *http.Request {
	r := defaultRequest()
	r.Header.Set("Sec-WebSocket-Version", "12")
	return r
}

func withIncorrectUpgrade() *http.Request {
	r := defaultRequest()
	r.Header.Set("Upgrade", "websockett")
	return r
}

func withIncorrectConnection() *http.Request {
	r := defaultRequest()
	r.Header.Set("Connection", "update")
	return r
}

func withIncorrectMethod() *http.Request {
	r := defaultRequest()
	r.Method = "POST"
	return r
}

func withNoSecKey() *http.Request {
	r := defaultRequest()
	r.Header.Set("Sec-WebSocket-Key", "")
	return r
}

func withNoHost() *http.Request {
	r := defaultRequest()
	r.Host = ""
	return r
}

func withNoRequestURI() *http.Request {
	r := defaultRequest()
	r.RequestURI = ""
	return r
}

func getRequest(req *http.Request) *http.Response {
	var ws WebSocket
	w := httptest.NewRecorder()
	ws.Handshake(w, req)
	resp := w.Result()
	return resp
}
