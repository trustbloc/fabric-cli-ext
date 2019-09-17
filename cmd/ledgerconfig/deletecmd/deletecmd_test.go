/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package deletecmd

import (
	"io"
	"testing"

	"github.com/hyperledger/fabric-cli/pkg/environment"
	"github.com/hyperledger/fabric-cli/pkg/fabric"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-cli-ext/cmd/ledgerconfig/common"
	"github.com/trustbloc/fabric-cli-ext/cmd/ledgerconfig/mocks"
)

func TestQueryCmd_InitializeError(t *testing.T) {
	t.Run("With channel error", func(t *testing.T) {
		errExpected := errors.New("channel error")
		p := func(config *environment.Config) (fabric.Factory, error) { return nil, errExpected }
		c := newMockCmd(t, p, "--criteria", `{"MspID":"msp1"}`, "--noprompt")
		require.EqualError(t, c.Execute(), errExpected.Error())
	})
	t.Run("With channel execute error", func(t *testing.T) {
		factory := &mocks.Factory{}
		ch := &mocks.Channel{}
		factory.ChannelReturns(ch, nil)

		errExpected := errors.New("channel execute error")
		ch.ExecuteReturns(channel.Response{}, errExpected)

		p := func(config *environment.Config) (fabric.Factory, error) { return factory, nil }
		c := newMockCmd(t, p, "--criteria", `{"MspID":"msp1"}`, "--noprompt")
		require.EqualError(t, c.Execute(), errExpected.Error())
	})
}

func TestDeleteCmd(t *testing.T) {
	factory := &mocks.Factory{}
	c := &mocks.Channel{}
	factory.ChannelReturns(c, nil)
	p := func(config *environment.Config) (fabric.Factory, error) { return factory, nil }

	t.Run("With --criteria", func(t *testing.T) {
		w := &mocks.Writer{}
		c := newMockCmdWithReaderWriter(t, &mocks.Reader{}, w, p, "--criteria", `{"MspID":"msp1"}`, "--noprompt")
		require.NoError(t, c.Execute())
		require.NotContains(t, w.Written(), msgContinueOrAbort) // Confirmation prompt should not be displayed
	})
	t.Run("With --mspid", func(t *testing.T) {
		w := &mocks.Writer{}
		c := newMockCmdWithReaderWriter(t, &mocks.Reader{}, w, p, "--mspid", "msp1", "--noprompt")
		require.NoError(t, c.Execute())
		require.NotContains(t, w.Written(), msgContinueOrAbort) // Confirmation prompt should not be displayed
	})
	t.Run("With prompt - Y", func(t *testing.T) {
		const payload = `[{"MspID":"msp1","PeerID":"","AppName":"app3","AppVersion":"1","TxID":"tx1","Format":"Other","Config":"config"}]`
		c.QueryReturns(channel.Response{Payload: []byte(payload)}, nil)
		w := &mocks.Writer{}
		c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("Y\n")}, w, p, "--criteria", `{"MspID":"msp1"}`)
		require.NoError(t, c.Execute())
		require.Contains(t, w.Written(), payload)            // The configuration to be deleted should be displayed
		require.Contains(t, w.Written(), msgContinueOrAbort) // The Y|N prompt should be displayed
		require.Contains(t, w.Written(), msgConfigDeleted)   // The message saying that the delete was successful should be displayed
	})
	t.Run("With prompt - N", func(t *testing.T) {
		const payload = `[{"MspID":"msp1","PeerID":"","AppName":"app3","AppVersion":"1","TxID":"tx1","Format":"Other","Config":"config"}]`
		c.QueryReturns(channel.Response{Payload: []byte(payload)}, nil)
		w := &mocks.Writer{}
		c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("N\n")}, w, p, "--criteria", `{"MspID":"msp1"}`)
		require.NoError(t, c.Execute())
		require.Contains(t, w.Written(), payload)             // The configuration to be deleted should be displayed
		require.Contains(t, w.Written(), msgContinueOrAbort)  // The Y|N prompt should be displayed
		require.Contains(t, w.Written(), msgAborted)          // The message saying that the delete was aborted should be displayed
		require.NotContains(t, w.Written(), msgConfigDeleted) // The message saying that the delete was successful should NOT be displayed
	})
	t.Run("With prompt - No config for criteria", func(t *testing.T) {
		c.QueryReturns(channel.Response{Payload: []byte("null")}, nil)
		w := &mocks.Writer{}
		c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("Y\n")}, w, p, "--mspid", "msp1")
		require.NoError(t, c.Execute())
		require.Equal(t, msgNoConfig, w.Written())
	})
	t.Run("With prompt - query error", func(t *testing.T) {
		errExpected := errors.New("query error")
		c.QueryReturns(channel.Response{}, errExpected)
		w := &mocks.Writer{}
		c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("Y\n")}, w, p, "--criteria", `{"MspID":"msp1"}`)
		require.EqualError(t, c.Execute(), errExpected.Error())
	})
}

func newMockCmd(t *testing.T, p common.FactoryProvider, args ...string) *cobra.Command {
	return newMockCmdWithReaderWriter(t, &mocks.Reader{}, &mocks.Writer{}, p, args...)
}

func newMockCmdWithReaderWriter(t *testing.T, in io.Reader, w io.Writer, p common.FactoryProvider, args ...string) *cobra.Command {
	settings := environment.NewDefaultSettings()
	settings.Streams.Out = w
	settings.Streams.In = in

	settings.Config.CurrentContext = "testctx"
	settings.Config.Contexts[settings.Config.CurrentContext] = &environment.Context{}

	c := newCmd(settings, p)
	require.NotNil(t, c)

	c.SetArgs(args)

	return c
}
