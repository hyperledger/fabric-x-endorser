/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package internal

import "runtime"

var (
	Version   = "Unknown"
	GoVersion = runtime.Version()
	Commit    = "Unknown"
	Os        = runtime.GOOS
	Arch      = runtime.GOARCH
)
