/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	jsonStr = `{"MspID":"Org1MSP","Apps":[{"AppName":"app1","Version":"1"}]}`

	formattedJSONStr = `{
  "MspID": "Org1MSP",
  "Apps": [
    {
      "AppName": "app1",
      "Version": "1"
    }
  ]
}`
)

func TestFormatJSON(t *testing.T) {
	formatted, err := FormatJSON([]byte(jsonStr))
	require.NoError(t, err)
	require.Equal(t, formattedJSONStr, string(formatted))
}
