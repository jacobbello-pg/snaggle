// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package netpolicy

import (
	"context"
	"errors"
	"net/netip"
	"testing"
)

type stubResolver struct {
	addrs []netip.Addr
	err   error
}

func (s stubResolver) LookupNetIP(context.Context, string, string) ([]netip.Addr, error) {
	return s.addrs, s.err
}

func TestCheckURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		stub stubResolver
		ok   bool
	}{
		{"public", "https://example.com", stubResolver{addrs: []netip.Addr{netip.MustParseAddr("93.184.216.34")}}, true},
		{"private DNS", "https://example.com", stubResolver{addrs: []netip.Addr{netip.MustParseAddr("10.0.0.1")}}, false},
		{"mixed DNS", "https://example.com", stubResolver{addrs: []netip.Addr{netip.MustParseAddr("93.184.216.34"), netip.MustParseAddr("127.0.0.1")}}, false},
		{"loopback literal", "http://127.0.0.1", stubResolver{}, false},
		{"shared address", "http://100.64.0.1", stubResolver{}, false},
		{"localhost", "http://localhost", stubResolver{}, false},
		{"local suffix", "http://service.internal", stubResolver{}, false},
		{"file", "file:///etc/passwd", stubResolver{}, false},
		{"credentials", "https://user:pass@example.com", stubResolver{}, false},
		{"DNS failure", "https://example.com", stubResolver{err: errors.New("offline")}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckURL(context.Background(), tt.stub, tt.url)
			if (err == nil) != tt.ok {
				t.Fatalf("CheckURL() error = %v, want success %v", err, tt.ok)
			}
		})
	}
}
