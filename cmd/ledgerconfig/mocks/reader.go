/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

// Reader provides a byte array from which the reader reads
type Reader struct {
	Bytes []byte
}

// Read reads from the byte array
func (r *Reader) Read(p []byte) (n int, err error) {
	return copy(p, r.Bytes), nil
}
