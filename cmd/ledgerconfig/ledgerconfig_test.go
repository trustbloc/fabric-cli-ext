/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"testing"

	"github.com/hyperledger/fabric-cli/pkg/environment"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-cli-ext/cmd/mocks"
)

func TestNew(t *testing.T) {
	settings := environment.NewDefaultSettings()

	cmd := New(settings)
	require.NotNil(t, cmd)

	w := &mocks.Writer{}
	cmd.SetOutput(w)

	err := cmd.Execute()
	require.NoError(t, err)

	// Make sure that the query command was added
	require.Contains(t, w.Written(), "Query ledger configuration")
	// Make sure that the update command was added
	require.Contains(t, w.Written(), "Update ledger configuration")
	// Make sure that the delete command was added
	require.Contains(t, w.Written(), "Delete ledger configuration")
	// Make sure that the fileidxupdate command was added
	require.Contains(t, w.Written(), "fileidxupdate")
}
