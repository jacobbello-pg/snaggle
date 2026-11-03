// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package browser

import "fmt"

// KillBrowser is disabled because a port number does not prove process
// ownership. The secure fork never sends signals to unrelated processes.
func (bm *BrowserManager) KillBrowser(port int) (int, error) {
	return 0, fmt.Errorf("browser process killing is disabled in the secure fork")
}
