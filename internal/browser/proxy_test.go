// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package browser

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestEgressProxyBlocksPrivateHTTP(t *testing.T) {
	proxy, err := startEgressProxy()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(proxy.Close)

	client := &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(&url.URL{Scheme: "http", Host: proxy.address()})},
		Timeout:   2 * time.Second,
	}
	resp, err := client.Get("http://127.0.0.1:1/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestEgressProxyBlocksPrivateConnect(t *testing.T) {
	proxy, err := startEgressProxy()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(proxy.Close)

	conn, err := net.DialTimeout("tcp", proxy.address(), 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_, _ = fmt.Fprint(conn, "CONNECT 127.0.0.1:443 HTTP/1.1\r\nHost: 127.0.0.1:443\r\n\r\n")
	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: http.MethodConnect})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestEgressProxyBlocksBrowserVendorBackground(t *testing.T) {
	proxy := &egressProxy{}
	if proxy.allowBrowserVendorHost("www.google.com") {
		t.Fatal("Chrome startup vendor host was allowed")
	}
	proxy.setPrimaryURL("https://example.com")
	if proxy.allowBrowserVendorHost("www.google.com") {
		t.Fatal("Chrome vendor background host was allowed after navigation")
	}
	if !proxy.allowBrowserVendorHost("fonts.googleapis.com") {
		t.Fatal("ordinary public page subresource was blocked")
	}
	proxy.setPrimaryURL("https://google.com")
	if !proxy.allowBrowserVendorHost("www.google.com") {
		t.Fatal("explicit target subdomain was blocked")
	}
}
