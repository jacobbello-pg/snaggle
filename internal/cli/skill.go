// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/p3bot/snag/internal/skill"
)

func skillFlagsChanged(cmd *cobra.Command) bool {
	return skillPrint
}

func handleSkill(cmd *cobra.Command, args []string) error {
	for _, arg := range args {
		if strings.TrimSpace(arg) != "" {
			return fmt.Errorf("--skill cannot be combined with URL arguments")
		}
	}
	_, err := fmt.Fprint(cmd.OutOrStdout(), skill.Text())
	return err
}
