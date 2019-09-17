/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package updatecmd

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

func TestUpdateCmd_InvalidOptions(t *testing.T) {
	t.Run("No options", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, nil).Execute(), errConfigOrConfigFileRequired)
	})

	t.Run("--config with --configfile", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, nil, "--config", "{}", "--configfile", "./config.json").Execute(), errConfigOrConfigFileRequired)
	})

	t.Run("Invalid config in --config flag", func(t *testing.T) {
		err := newMockCmd(t, nil, "--config", "invalid-JSON").Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), errInvalidJSONConfig)
	})

	t.Run("File in --configfile flag not found", func(t *testing.T) {
		err := newMockCmd(t, nil, "--configfile", "./notthere.json").Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), errFileNotFound)
	})
}

func TestUpdateCmd_InitializeError(t *testing.T) {
	t.Run("With channel error", func(t *testing.T) {
		errExpected := errors.New("channel error")
		p := func(config *environment.Config) (fabric.Factory, error) { return nil, errExpected }
		c := newMockCmd(t, p, "--config", `{"MspID":"msp1"}`, "--noprompt")
		require.EqualError(t, c.Execute(), errExpected.Error())
	})
	t.Run("With channel execute error", func(t *testing.T) {
		factory := &mocks.Factory{}
		ch := &mocks.Channel{}
		factory.ChannelReturns(ch, nil)

		errExpected := errors.New("channel execute error")
		ch.ExecuteReturns(channel.Response{}, errExpected)

		p := func(config *environment.Config) (fabric.Factory, error) { return factory, nil }
		c := newMockCmd(t, p, "--config", `{"MspID":"msp1"}`, "--noprompt")
		require.EqualError(t, c.Execute(), errExpected.Error())
	})
}

func TestUpdateCmd(t *testing.T) {
	factory := &mocks.Factory{}
	c := &mocks.Channel{}

	resp := channel.Response{}
	c.QueryReturns(resp, nil)
	factory.ChannelReturns(c, nil)

	p := func(config *environment.Config) (fabric.Factory, error) { return factory, nil }

	t.Run("With --config", func(t *testing.T) {
		c := newMockCmd(t, p, "--config", `{"MspID":"msp1"}`, "--noprompt")
		require.NoError(t, c.Execute())
	})

	t.Run("With --file", func(t *testing.T) {
		w := &mocks.Writer{}
		c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("Y\n")}, w, p, "--configfile", "../sampleconfig/org1-config.json", "--noprompt")
		require.NoError(t, c.Execute())
		require.Contains(t, w.Written(), msgConfigUpdated)
	})

	t.Run("With --file and relative file references - ref file not found", func(t *testing.T) {
		c := newMockCmd(t, p, "--configfile", "../sampleconfig/invalid-refs-config.json", "--noprompt")
		err := c.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "../sampleconfig/file-not-there.json: no such file or directory")
	})

	t.Run("With --file and absolute file references - ref file not found", func(t *testing.T) {
		c := newMockCmd(t, p, "--configfile", "../sampleconfig/absolute-refs-config.json", "--noprompt")
		err := c.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "/usr/local/backup/file-not-there.json: no such file or directory")
	})

	t.Run("With prompt - Y", func(t *testing.T) {
		w := &mocks.Writer{}
		c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("Y\n")}, w, p, "--config", `{"MspID":"msp1"}`)
		require.NoError(t, c.Execute())
		require.Contains(t, w.Written(), msgContinueOrAbort)
		require.Contains(t, w.Written(), msgConfigUpdated)
	})

	t.Run("With prompt - N", func(t *testing.T) {
		w := &mocks.Writer{}
		c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("N\n")}, w, p, "--config", `{"MspID":"msp1"}`)
		require.NoError(t, c.Execute())
		require.Contains(t, w.Written(), msgContinueOrAbort)
		require.Contains(t, w.Written(), msgAborted)
		require.NotContains(t, w.Written(), msgConfigUpdated)
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
