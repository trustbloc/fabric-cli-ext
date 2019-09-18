/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"bytes"
	"encoding/json"
)

// FormatJSON transforms the given JSON into a displayable format
func FormatJSON(jsonBytes []byte) ([]byte, error) {
	var buff bytes.Buffer
	err := json.Indent(&buff, jsonBytes, "", "  ")
	if err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}
