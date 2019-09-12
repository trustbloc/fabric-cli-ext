/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"testing"

	"github.com/hyperledger/fabric-cli/pkg/environment"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	cmd := New(environment.NewDefaultSettings())
	require.NotNil(t, cmd)

	err := cmd.Execute()
	require.NoError(t, err)
}
