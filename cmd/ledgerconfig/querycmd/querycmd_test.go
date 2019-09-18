/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package querycmd

import (
	"errors"
	"io"
	"testing"

	"github.com/hyperledger/fabric-cli/pkg/environment"
	"github.com/hyperledger/fabric-cli/pkg/fabric"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-cli-ext/cmd/ledgerconfig/common"
	"github.com/trustbloc/fabric-cli-ext/cmd/ledgerconfig/mocks"
)

const (
	payload = `[{"MspID":"msp1"}]`

	formattedPayload = `[
  {
    "MspID": "msp1"
  }
]
`
)

func TestQueryCmd_InitializeError(t *testing.T) {
	t.Run("With channel error", func(t *testing.T) {
		errExpected := errors.New("channel error")
		p := func(config *environment.Config) (fabric.Factory, error) { return nil, errExpected }
		c := newMockCmd(t, &mocks.Writer{}, p, "--criteria", `{"MspID":"msp1"}`)
		require.EqualError(t, c.Execute(), errExpected.Error())
	})
	t.Run("With channel query error", func(t *testing.T) {
		factory := &mocks.Factory{}
		ch := &mocks.Channel{}
		factory.ChannelReturns(ch, nil)

		errExpected := errors.New("channel query error")
		ch.QueryReturns(channel.Response{}, errExpected)

		p := func(config *environment.Config) (fabric.Factory, error) { return factory, nil }
		c := newMockCmd(t, &mocks.Writer{}, p, "--criteria", `{"MspID":"msp1"}`)
		require.EqualError(t, c.Execute(), errExpected.Error())
	})
}

func TestQueryCmd(t *testing.T) {
	factory := &mocks.Factory{}
	c := &mocks.Channel{}

	resp := channel.Response{
		Payload: []byte(payload),
	}
	c.QueryReturns(resp, nil)
	factory.ChannelReturns(c, nil)

	p := func(config *environment.Config) (fabric.Factory, error) { return factory, nil }

	t.Run("With --criteria", func(t *testing.T) {
		w := &mocks.Writer{}
		c := newMockCmd(t, w, p, "--criteria", `{"MspID":"msp1"}`)
		require.NoError(t, c.Execute())
		require.Equal(t, payload, w.Written())
	})
	t.Run("With --mspid", func(t *testing.T) {
		w := &mocks.Writer{}
		c := newMockCmd(t, w, p, "--mspid", "msp1")
		require.NoError(t, c.Execute())
		require.Equal(t, payload, w.Written())
	})
	t.Run("No config for criteria", func(t *testing.T) {
		c.QueryReturns(channel.Response{Payload: []byte("null")}, nil)
		w := &mocks.Writer{}
		c := newMockCmd(t, w, p, "--mspid", "msp1")
		require.NoError(t, c.Execute())
		require.Equal(t, msgNoConfig, w.Written())
	})
	t.Run("With --format", func(t *testing.T) {
		c.QueryReturns(resp, nil)
		w := &mocks.Writer{}
		c := newMockCmd(t, w, p, "--mspid", "msp1", "--format")
		require.NoError(t, c.Execute())
		require.Equal(t, formattedPayload, string(w.Bytes))
	})
	t.Run("With format error", func(t *testing.T) {
		c.QueryReturns(channel.Response{Payload: []byte("invalid JSON")}, nil)
		c := newMockCmd(t, &mocks.Writer{}, p, "--mspid", "msp1", "--format")
		require.Error(t, c.Execute())
	})
}

func newMockCmd(t *testing.T, out io.Writer, p common.FactoryProvider, args ...string) *cobra.Command {
	settings := environment.NewDefaultSettings()
	settings.Streams.Out = out

	settings.Config.CurrentContext = "testctx"
	settings.Config.Contexts[settings.Config.CurrentContext] = &environment.Context{}

	c := newCmd(settings, p)
	require.NotNil(t, c)

	c.SetArgs(args)

	return c
}
