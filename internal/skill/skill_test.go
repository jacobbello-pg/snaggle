// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package skill_test

import (
	"strings"
	"testing"

	"github.com/p3bot/snag/internal/skill"
)

func TestTextFrontmatterAndH1(t *testing.T) {
	text := skill.Text()
	if !strings.HasPrefix(text, "---\nname: snaggle\n") {
		t.Fatal("skill must open with Agent Skills frontmatter (name: snaggle)")
	}
	if !strings.Contains(text, "description:") {
		t.Fatal("skill frontmatter must include description")
	}
	if !strings.Contains(text, "\n# Use Snaggle as a secondary web fetcher\n") {
		t.Fatal("skill must have H1 after frontmatter")
	}
	if !strings.HasSuffix(text, "\n") {
		t.Fatal("Text must end with a newline")
	}
}

func TestTextRequiredGuidance(t *testing.T) {
	text := skill.Text()
	needles := []string{
		"secondary web fetcher",
		"ordinary web tool fails",
		"public-network-only",
		"ephemeral profile",
		"Do not use it for authenticated",
		"Do not use `--trusted-session`",
		"snag --timeout 30 --format md https://example.com",
	}
	for _, n := range needles {
		if !strings.Contains(text, n) {
			t.Errorf("skill missing required guidance %q", n)
		}
	}
}
