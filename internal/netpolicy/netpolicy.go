// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package netpolicy restricts browser traffic to public HTTP(S) destinations.
package netpolicy

import (
	"context"
	"fmt"
	"net/netip"
	"net/url"
	"strings"
)

// Resolver is the DNS surface used by CheckURL.
type Resolver interface {
	LookupNetIP(context.Context, string, string) ([]netip.Addr, error)
}

// ResolveURL validates a URL and returns every public address it resolves to.
func ResolveURL(ctx context.Context, resolver Resolver, rawURL string) ([]netip.Addr, error) {
	u, err := checkURLSyntax(rawURL)
	if err != nil {
		return nil, err
	}
	host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	if addr, err := netip.ParseAddr(host); err == nil {
		if err := checkAddr(addr); err != nil {
			return nil, err
		}
		return []netip.Addr{addr.Unmap()}, nil
	}

	addrs, err := resolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return nil, fmt.Errorf("resolve %s: %w", host, err)
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("resolve %s: no addresses", host)
	}
	for _, addr := range addrs {
		if err := checkAddr(addr); err != nil {
			return nil, fmt.Errorf("destination %s: %w", host, err)
		}
	}
	return addrs, nil
}

// CheckURL rejects local files, credentials, local names, and destinations
// resolving to non-public address space. Every resolved address must be public.
func CheckURL(ctx context.Context, resolver Resolver, rawURL string) error {
	_, err := ResolveURL(ctx, resolver, rawURL)
	return err
}

// CheckURLSyntax rejects unsafe schemes, embedded credentials, local names,
// and non-public literal IP addresses without performing DNS.
func CheckURLSyntax(rawURL string) error {
	_, err := checkURLSyntax(rawURL)
	return err
}

func checkURLSyntax(rawURL string) (*url.URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("only public HTTP(S) URLs are allowed")
	}
	if u.User != nil {
		return nil, fmt.Errorf("URL credentials are not allowed")
	}
	host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	if host == "" || localName(host) {
		return nil, fmt.Errorf("local destination is not allowed")
	}
	if addr, err := netip.ParseAddr(host); err == nil {
		if err := checkAddr(addr); err != nil {
			return nil, err
		}
	}
	return u, nil
}

func localName(host string) bool {
	return host == "localhost" ||
		strings.HasSuffix(host, ".localhost") ||
		strings.HasSuffix(host, ".local") ||
		strings.HasSuffix(host, ".internal") ||
		strings.HasSuffix(host, ".home.arpa")
}

func checkAddr(addr netip.Addr) error {
	addr = addr.Unmap()
	if !addr.IsValid() || !addr.IsGlobalUnicast() || addr.IsPrivate() ||
		addr.IsLoopback() || addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() ||
		addr.IsMulticast() || addr.IsUnspecified() || sharedIPv4.Contains(addr) {
		return fmt.Errorf("non-public address %s is not allowed", addr)
	}
	return nil
}

var sharedIPv4 = netip.MustParsePrefix("100.64.0.0/10")
