/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package types

import (
	"google.golang.org/protobuf/encoding/protowire"
)

type (
	CommitterVersion string
	// VersionNumber represents a row's version.
	VersionNumber uint64
)

// Bytes converts a version number representation to bytes representation.
func (v VersionNumber) Bytes() []byte {
	return protowire.AppendVarint(nil, uint64(v))
}

// VersionNumberFromBytes converts a version bytes representation to a number representation.
func VersionNumberFromBytes(version []byte) VersionNumber {
	v, _ := protowire.ConsumeVarint(version)
	return VersionNumber(v)
}
