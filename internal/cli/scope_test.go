// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnforceTrustBoundary(t *testing.T) {
	resetCLIFlags()
	t.Cleanup(resetCLIFlags)

	if err := rootCmd.Flags().Set("output", "page.md"); err != nil {
		t.Fatal(err)
	}
	if err := enforceTrustBoundary(rootCmd); err != nil {
		t.Fatalf("ordinary output rejected: %v", err)
	}

	if err := rootCmd.Flags().Set("tab", "1"); err != nil {
		t.Fatal(err)
	}
	if err := enforceTrustBoundary(rootCmd); err == nil {
		t.Fatal("existing-tab access must require trusted mode")
	}

	trustedSession = true
	if err := enforceTrustBoundary(rootCmd); err != nil {
		t.Fatalf("trusted tab access rejected: %v", err)
	}
}

func TestBrowserOptionsNormalModeIsIsolated(t *testing.T) {
	resetCLIFlags()
	t.Cleanup(resetCLIFlags)

	opts, err := browserOptionsFromFlags(rootCmd, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if !opts.ForceHeadless || opts.OpenBrowser || !opts.TempProfile || opts.Port != 0 {
		t.Fatalf("normal options are not isolated: %+v", opts)
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

	outDir := t.TempDir()
	pngPath := filepath.Join(outDir, "page.png")
	_, stderr, err = runSnag("--format", "png", "--output", pngPath, "https://example.com")
	if err != nil {
		t.Fatalf("PNG fetch: %v\nstderr=%s", err, stderr)
	}
	pngData, err := os.ReadFile(pngPath)
	if err != nil || !bytes.HasPrefix(pngData, []byte("\x89PNG\r\n\x1a\n")) {
		t.Fatalf("invalid PNG output: err=%v bytes=%d", err, len(pngData))
	}

	pdfPath := filepath.Join(outDir, "page.pdf")
	_, stderr, err = runSnag("--format", "pdf", "--output", pdfPath, "https://example.com")
	if err != nil {
		t.Fatalf("PDF fetch: %v\nstderr=%s", err, stderr)
	}
	pdfData, err := os.ReadFile(pdfPath)
	if err != nil || !bytes.HasPrefix(pdfData, []byte("%PDF-")) {
		t.Fatalf("invalid PDF output: err=%v bytes=%d", err, len(pdfData))
	}

	batchDir := filepath.Join(outDir, "batch")
	if err := os.Mkdir(batchDir, 0o755); err != nil {
		t.Fatal(err)
	}
	_, stderr, err = runSnag("--output-dir", batchDir,
		"https://example.com", "https://example.com/?second=1")
	if err != nil {
		t.Fatalf("batch fetch: %v\nstderr=%s", err, stderr)
	}
	entries, err := os.ReadDir(batchDir)
	if err != nil || len(entries) != 2 {
		t.Fatalf("batch output: err=%v files=%d", err, len(entries))
	}

	stdout, stderr, err = runSnag("http://127.0.0.1:1")
	if err == nil {
		t.Fatalf("private target unexpectedly succeeded: %s", stdout)
	}
	if !strings.Contains(stderr, "non-public") {
		t.Fatalf("private target error = %q", stderr)
	}
}
