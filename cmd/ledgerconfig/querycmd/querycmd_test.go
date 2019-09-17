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

func TestQueryCmd_InvalidOptions(t *testing.T) {
	t.Run("No options", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, &mocks.Writer{}, nil).Execute(), errMspOrCriteriaRequired)
	})

	t.Run("--mspid with --criteria", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, &mocks.Writer{}, nil, "--mspid", "MSP1", "--criteria", "{}").Execute(), errCriteriaMustBeAlone)
	})

	t.Run("--peerid with --criteria", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, &mocks.Writer{}, nil, "--peerid", "peer1", "--criteria", "{}").Execute(), errCriteriaMustBeAlone)
	})

	t.Run("--appname with --criteria", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, &mocks.Writer{}, nil, "--appname", "app1", "--criteria", "{}").Execute(), errCriteriaMustBeAlone)
	})

	t.Run("--appver with --criteria", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, &mocks.Writer{}, nil, "--appver", "v1", "--criteria", "{}").Execute(), errCriteriaMustBeAlone)
	})

	t.Run("--componentname with --criteria", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, &mocks.Writer{}, nil, "--componentname", "comp1", "--criteria", "{}").Execute(), errCriteriaMustBeAlone)
	})

	t.Run("--componentver with --criteria", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, &mocks.Writer{}, nil, "--componentver", "v1", "--criteria", "{}").Execute(), errCriteriaMustBeAlone)
	})

	t.Run("Invalid --criteria", func(t *testing.T) {
		c := newMockCmd(t, &mocks.Writer{}, nil, "--criteria", "xxx")
		err := c.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), errInvalidCriteria)
	})
}

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

	const payload = `[{"MspID":"msp1"}]`
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
