// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"github.com/p3bot/snag/internal/logger"
	"github.com/spf13/pflag"
)

func resetCLIFlags() {
	urlFile, flagOutput, outputDir, flagFormat = "", "", "", "md"
	timeout, waitFor, port = DefaultTimeout, "", 0
	closeTab, forceHead, openBrowser = false, false, false
	listTabs, allTabs, killBrowser, runDoctor = false, false, false, false
	tab, userAgent, userDataDir = "", "", ""
	showVersion, info, verbose, debug, tempProfile, skillPrint = false, false, false, false, false, false
	colorMode = logger.ColorAuto
	reset := func(fs *pflag.FlagSet) {
		fs.VisitAll(func(f *pflag.Flag) { f.Changed = false })
	}
	reset(rootCmd.Flags())
	reset(rootCmd.PersistentFlags())
}
