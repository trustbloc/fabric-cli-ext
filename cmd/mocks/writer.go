/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"strings"
)

// Writer captures the written bytes so that they may be examined in unit tests
type Writer struct {
	Bytes []byte
	Err   error
}

// Write saves the given bytes
func (w *Writer) Write(p []byte) (int, error) {
	if w.Err != nil {
		return 0, w.Err
	}
	w.Bytes = append(w.Bytes, p...)
	return len(p), nil
}

// Written returns the written bytes as a string (minus any newline characters)
func (w *Writer) Written() string {
	// Remove all newline chars
	return strings.Replace(string(w.Bytes), "\n", "", -1)
}
