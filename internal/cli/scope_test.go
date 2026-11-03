// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import "testing"

func TestEnforceSecondaryScope(t *testing.T) {
	resetCLIFlags()
	t.Cleanup(resetCLIFlags)

	if err := enforceSecondaryScope(rootCmd, []string{"https://example.com"}); err != nil {
		t.Fatalf("safe invocation rejected: %v", err)
	}
	if err := enforceSecondaryScope(rootCmd, []string{"https://a.example", "https://b.example"}); err == nil {
		t.Fatal("multiple URLs must be rejected")
	}

	flagFormat = "pdf"
	if err := enforceSecondaryScope(rootCmd, []string{"https://example.com"}); err == nil {
		t.Fatal("binary output must be rejected")
	}
	flagFormat = "md"

	if err := rootCmd.Flags().Set("output", "page.md"); err != nil {
		t.Fatal(err)
	}
	if err := enforceSecondaryScope(rootCmd, []string{"https://example.com"}); err == nil {
		t.Fatal("file output must be rejected")
	}
}
