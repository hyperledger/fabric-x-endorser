/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package utils

func Map[I, O any](ins []I, f func(I) O) []O {
	if ins == nil {
		return nil
	}
	outs := make([]O, len(ins))
	for i, in := range ins {
		outs[i] = f(in)
	}
	return outs
}
