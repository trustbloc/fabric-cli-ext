/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package instantiatecmd

import (
	"io"
	"testing"

	"github.com/hyperledger/fabric-cli/pkg/environment"
	"github.com/hyperledger/fabric-cli/pkg/fabric"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/fabric-cli-ext/cmd/basecmd"
	"github.com/trustbloc/fabric-cli-ext/cmd/mocks"
)

func TestInstantiateCmd_InitializeError(t *testing.T) {
}

func TestInstantiateCmd(t *testing.T) {
	factory := &mocks.Factory{}
	r := &mocks.ResMgmt{}

	factory.ResourceManagementReturns(r, nil)

	p := func(config *environment.Config) (fabric.Factory, error) { return factory, nil }

	t.Run("Missing chaincode name arg", func(t *testing.T) {
		w := &mocks.Writer{}
		c := newMockCmd(t, w, p)
		require.EqualError(t, c.Execute(), "chaincode name not specified")
		require.Equal(t, "Error: chaincode name not specified", w.Written())
	})

	t.Run("Missing chaincode version arg", func(t *testing.T) {
		w := &mocks.Writer{}
		c := newMockCmd(t, w, p, "cc1")
		require.EqualError(t, c.Execute(), "chaincode version not specified")
		require.Equal(t, "Error: chaincode version not specified", w.Written())
	})

	t.Run("Success", func(t *testing.T) {
		w := &mocks.Writer{}
		c := newMockCmd(t, w, p, "cc1", "v1")
		require.NoError(t, c.Execute())
		require.Contains(t, w.Written(), msgCCInstantiated)
	})

	t.Run("With policy -> Success", func(t *testing.T) {
		w := &mocks.Writer{}
		c := newMockCmd(t, w, p, "cc1", "v1", "--policy", "OR('Org1.member','Org2.member')")
		require.NoError(t, c.Execute())
		require.Contains(t, w.Written(), msgCCInstantiated, w.Written())
	})

	t.Run("With invalid policy -> Success", func(t *testing.T) {
		w := &mocks.Writer{}
		c := newMockCmd(t, w, p, "cc1", "v1", "--policy", "OR('Org1.member','Org2')")
		require.EqualError(t, c.Execute(), "error parsing chaincode policy")
		require.Equal(t, "Error: error parsing chaincode policy", w.Written())
	})

	t.Run("With collections config -> Success", func(t *testing.T) {
		const collsCfg = `[{"name":"coll1","type":"COL_DCAS","policy":"OR('Org1MSP.member','Org2MSP.member')","maxPeerCount":2,"requiredPeerCount":1,"timeToLive":"10m"},{"name":"coll2","type":"COL_OFFLEDGER","policy":"OR('IMPLICIT-ORG.member')"}]`

		w := &mocks.Writer{}
		c := newMockCmd(t, w, p, "cc1", "v1", "--collections-config", collsCfg)
		require.NoError(t, c.Execute())
		require.Contains(t, w.Written(), msgCCInstantiated)
	})

	t.Run("With invalid collections config -> Success", func(t *testing.T) {
		const collsCfg = `{`

		w := &mocks.Writer{}
		c := newMockCmd(t, w, p, "cc1", "v1", "--collections-config", collsCfg)
		err := c.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid collections config")
		require.Contains(t, w.Written(), "Error: invalid collections config")
	})
}

func newMockCmd(t *testing.T, out io.Writer, p basecmd.FactoryProvider, args ...string) *cobra.Command {
	settings := environment.NewDefaultSettings()
	settings.Streams.Out = out

	settings.Config.CurrentContext = "testctx"
	settings.Config.Contexts[settings.Config.CurrentContext] = &environment.Context{}

	c := newCmd(settings, p)
	require.NotNil(t, c)

	c.SetArgs(args)

	return c
}
