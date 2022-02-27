package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
)

const magic = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

func main() {
	ws := WebSocket{}
	http.HandleFunc("/websocket", ws.Handshake) // TODO(ben): We will need to be aware of the state of this protocol. Was it successful? is it browser-based? What protocols & extensions?
	go func() {
		log.Fatal(http.ListenAndServe(":8000", nil))
	}()

	// send an HTTP request to that server.
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:8000/websocket", nil)
	if err != nil {
		panic(err)
	}

	// websocket handshake request
	req.Header.Add("Upgrade", "websocket")
	req.Header.Add("Connection", "Upgrade")
	req.Header.Add("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==") // This will be a random 16-byte base-64 encoded value.
	req.Header.Add("Sec-WebSocket-Version", "13")

	raw, _ := httputil.DumpRequestOut(req, false)
	fmt.Println(string(raw))

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// read the response
	dmp, _ := httputil.DumpResponse(resp, false)
	fmt.Printf("in main, dump response: %s", dmp)
}

type WebSocket struct {
	browser bool
	open    bool
	// protocol & extensions
}

// TODO(ben): attach body to status response
func (ws *WebSocket) Handshake(w http.ResponseWriter, r *http.Request) {
	// RFC 6455. Section 4.2.1: Reading the Client's Opening Handshakes
	version := r.Proto
	if version != "HTTP/1.1" && version != "HTTP/2.0" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	request := r.Method
	if request != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resource := r.RequestURI
	if resource == "" { // TODO(ben) handle 404 not found request uris.
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Host == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Header.Get("Upgrade") != "websocket" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Header.Get("Connection") != "Upgrade" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil || len(decoded) != 16 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Header.Get("Sec-Websocket-Version") != "13" {
		w.Header().Add("Sec-Websocket-Version", "13")
		w.WriteHeader(http.StatusUpgradeRequired)
		return
	}
	origin := r.Header.Get("Origin") // TODO(ben): Expose ways to validate origin?
	if origin != "" {
		ws.browser = true
	}

	// r.Header.Get("Sec-WebSocket-Protocol")
	// r.Header.Get("Sec-WebSocket-Extensions")

	// TODO(ben): handle TLS.

	// TODO(ben): Explore the usecase where we want to be able to redirect  or authenticate. How should the handshake handle those events?

	// Acknowledged request. Respond to  client.
	h := sha1.New()
	io.WriteString(h, key)
	io.WriteString(h, magic)
	sum := h.Sum(nil)
	enc := base64.StdEncoding.EncodeToString(sum)
	w.Header().Set("Sec-WebSocket-Accept", enc)
	w.Header().Set("Upgrade", "websocket")
	w.Header().Set("Connection", "Upgrade")
	// w.Header().Set("Sec-WebSocket-Protocol", "")
	// w.Header().Set("Sec-WebSocket-Extensions", "")

	w.WriteHeader(http.StatusSwitchingProtocols)
	ws.open = true
}
