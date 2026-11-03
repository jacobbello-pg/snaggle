// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"os"
	"strings"
	"testing"
)

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

func TestSecureCLIChromeIntegration(t *testing.T) {
	if os.Getenv("SNAG_BROWSER_INTEGRATION") != "1" {
		t.Skip("set SNAG_BROWSER_INTEGRATION=1 to launch isolated Chromium")
	}

	stdout, stderr, err := runSnag("--format", "text", "--timeout", "30", "https://example.com")
	if err != nil {
		t.Fatalf("public fetch: %v\nstderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, "This domain is for use in documentation examples") {
		t.Fatalf("public fetch missing expected content: %q", stdout)
	}

	stdout, stderr, err = runSnag("http://127.0.0.1:1")
	if err == nil {
		t.Fatalf("private target unexpectedly succeeded: %s", stdout)
	}
	if !strings.Contains(stderr, "non-public") {
		t.Fatalf("private target error = %q", stderr)
	}
}
