// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package browser

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/p3bot/snag/internal/netpolicy"
)

type egressProxy struct {
	listener  net.Listener
	server    *http.Server
	transport *http.Transport
	mu        sync.Mutex
	primary   string
	allowed   map[string]struct{}
	blocked   map[string]struct{}
}

func startEgressProxy() (*egressProxy, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("start browser egress proxy: %w", err)
	}
	p := &egressProxy{
		listener: listener,
		allowed:  make(map[string]struct{}),
		blocked:  make(map[string]struct{}),
	}
	p.transport = &http.Transport{
		Proxy:               nil,
		DialContext:         p.dialPublic,
		ForceAttemptHTTP2:   true,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	p.server = &http.Server{
		Handler:           p,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       30 * time.Second,
		ErrorLog:          log.New(io.Discard, "", 0),
	}
	go func() { _ = p.server.Serve(listener) }()
	return p, nil
}

func (p *egressProxy) address() string {
	return p.listener.Addr().String()
}

func (p *egressProxy) Close() {
	p.transport.CloseIdleConnections()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = p.server.Shutdown(ctx)
}

func (p *egressProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		p.connect(w, r)
		return
	}
	if r.URL == nil || r.URL.Scheme != "http" {
		p.recordHost(r.Host, false)
		http.Error(w, "public HTTP(S) only", http.StatusForbidden)
		return
	}
	if !p.allowBrowserVendorHost(r.URL.Hostname()) {
		p.recordHost(r.Host, false)
		http.Error(w, "browser background destination blocked", http.StatusForbidden)
		return
	}
	if err := netpolicy.CheckURLSyntax(r.URL.String()); err != nil {
		p.recordHost(r.Host, false)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	out := r.Clone(r.Context())
	out.RequestURI = ""
	out.Header.Del("Proxy-Authorization")
	out.Header.Del("Proxy-Connection")
	resp, err := p.transport.RoundTrip(out)
	if err != nil {
		p.recordHost(r.Host, false)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	p.recordHost(r.Host, true)
	defer resp.Body.Close()
	for k, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(k, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (p *egressProxy) recordHost(hostport string, allowed bool) {
	host := hostport
	if parsed, _, err := net.SplitHostPort(hostport); err == nil {
		host = parsed
	}
	host = strings.Trim(strings.ToLower(host), "[]")
	p.mu.Lock()
	if allowed {
		p.allowed[host] = struct{}{}
	} else {
		p.blocked[host] = struct{}{}
	}
	p.mu.Unlock()
}

func (p *egressProxy) observedHosts(allowed bool) []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	source := p.blocked
	if allowed {
		source = p.allowed
	}
	hosts := make([]string, 0, len(source))
	for host := range source {
		hosts = append(hosts, host)
	}
	return hosts
}

func (p *egressProxy) setPrimaryURL(rawURL string) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return
	}
	p.mu.Lock()
	p.primary = strings.ToLower(u.Hostname())
	p.mu.Unlock()
}

func (p *egressProxy) allowBrowserVendorHost(host string) bool {
	host = strings.ToLower(host)
	p.mu.Lock()
	primary := p.primary
	p.mu.Unlock()
	if host == primary || strings.HasSuffix(host, "."+primary) {
		return true
	}
	return !isBrowserVendorHost(host)
}

func isBrowserVendorHost(host string) bool {
	host = strings.ToLower(host)
	for _, blocked := range []string{
		"www.google.com",
		"accounts.google.com",
		"www.gstatic.com",
		"clients2.google.com",
		"clients4.google.com",
		"update.googleapis.com",
		"safebrowsing.googleapis.com",
		"optimizationguide-pa.googleapis.com",
	} {
		if host == blocked {
			return true
		}
	}
	return false
}

func (p *egressProxy) connect(w http.ResponseWriter, r *http.Request) {
	host, _, _ := net.SplitHostPort(r.Host)
	if !p.allowBrowserVendorHost(strings.Trim(host, "[]")) {
		p.recordHost(r.Host, false)
		http.Error(w, "browser background destination blocked", http.StatusForbidden)
		return
	}
	destination, err := p.dialPublic(r.Context(), "tcp", r.Host)
	if err != nil {
		p.recordHost(r.Host, false)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	p.recordHost(r.Host, true)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		destination.Close()
		http.Error(w, "proxy tunnel unavailable", http.StatusInternalServerError)
		return
	}
	client, buffered, err := hijacker.Hijack()
	if err != nil {
		destination.Close()
		return
	}
	_, _ = buffered.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n")
	_ = buffered.Flush()

	done := make(chan struct{}, 2)
	go copyTunnel(destination, client, done)
	go copyTunnel(client, destination, done)
	<-done
	_ = client.Close()
	_ = destination.Close()
}

func copyTunnel(dst net.Conn, src net.Conn, done chan<- struct{}) {
	_, _ = io.Copy(dst, src)
	done <- struct{}{}
}

func (p *egressProxy) dialPublic(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("invalid destination %q: %w", address, err)
	}
	host = strings.Trim(host, "[]")
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return nil, fmt.Errorf("invalid destination port %q", port)
	}
	checkURL := (&url.URL{Scheme: "https", Host: net.JoinHostPort(host, port)}).String()
	addrs, err := netpolicy.ResolveURL(ctx, net.DefaultResolver, checkURL)
	if err != nil {
		return nil, err
	}

	var lastErr error
	dialer := net.Dialer{Timeout: 10 * time.Second}
	for _, addr := range addrs {
		conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(addr.String(), port))
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("connect to public destination: %w", lastErr)
}
