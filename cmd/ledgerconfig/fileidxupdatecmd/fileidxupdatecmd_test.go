/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fileidxupdatecmd

import (
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/hyperledger/fabric-cli/pkg/environment"
	"github.com/hyperledger/fabric-cli/pkg/fabric"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-cli-ext/cmd/basecmd"
	"github.com/trustbloc/fabric-cli-ext/cmd/ledgerconfig/common"
	"github.com/trustbloc/fabric-cli-ext/cmd/mocks"
)

func TestNew(t *testing.T) {
	require.NotNil(t, New(environment.NewDefaultSettings()))
}

func TestFileIDXUpdateCmd_InvalidOptions(t *testing.T) {
	const (
		mspFlag   = "--msp"
		msp       = "Org1MSP"
		peersFlag = "--peers"
		peers     = "peer0.org1.example.com;peer1.org1.example.com"
		pathFlag  = "--path"
		path      = "/content"
	)

	t.Run("No options", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, nil).Execute(), errMSPRequired.Error())
	})

	t.Run("No peers", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, nil, mspFlag, msp).Execute(), errPeersRequired.Error())
	})

	t.Run("No path", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, nil, mspFlag, msp, peersFlag, peers).Execute(), errBasePathRequired.Error())
	})

	t.Run("No file index ID", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, nil, mspFlag, msp, peersFlag, peers, pathFlag, path).Execute(), errFileIndexIDRequired.Error())
	})
}

func TestFileIDXUpdateCmd_InitializeError(t *testing.T) {
	args := []string{"--msp", "Org1MSP", "--peers", "peer0.org1.example.com;peer1.org1.example.com", "--path", "/content", "--idxid", "file:idx:EiAuN66iEpuRt6IIu-2sO3bRM74sS_AIuY6jTbtFUsqAaA=="}

	t.Run("With channel error", func(t *testing.T) {
		errExpected := errors.New("channel error")
		p := func(config *environment.Config) (fabric.Factory, error) { return nil, errExpected }

		c := newMockCmd(t, p, append(args, "--noprompt")...)
		require.EqualError(t, c.Execute(), errExpected.Error())
	})
}

func TestFileIDXUpdateCmd(t *testing.T) {
	const (
		msp  = "Org1MSP"
		peer = "peer.org1.example.com"
		path = "/content"

		handlerCfg           = `{"BasePath":"/content","ChaincodeName":"files","Collection":"consortium","IndexNamespace":"file:idx"}`
		mismatchedHandlerCfg = `{"BasePath":"/content","ChaincodeName":"files","Collection":"consortium","IndexNamespace":"file:xxx"}`
		setHandlerCfg        = `{"BasePath":"/content","ChaincodeName":"files","Collection":"consortium","IndexNamespace":"file:idx","IndexDocID": "file:idx:EiAuN66iEpuRt6IIu-2sO3bRM74sS_AIuY6jTbtFUsqAaA=="}`
	)

	args := []string{"--msp", "Org1MSP", "--peers", "peer.org1.example.com;peer1.org1.example.com", "--path", "/content", "--idxid", "file:idx:EiAuN66iEpuRt6IIu-2sO3bRM74sS_AIuY6jTbtFUsqAaA=="}

	factory := &mocks.Factory{}
	c := &mocks.Channel{}

	factory.ChannelReturns(c, nil)

	p := func(config *environment.Config) (fabric.Factory, error) { return factory, nil }

	key := &common.Key{
		MspID:            msp,
		PeerID:           peer,
		AppName:          fileHandlerAppName,
		AppVersion:       fileHandlerAppVersion,
		ComponentName:    path,
		ComponentVersion: "1",
	}

	validCfg := &common.KeyValue{
		Key:   key,
		Value: &common.Value{TxID: "tx1", Format: "json", Config: handlerCfg},
	}

	validCfgBytes, err := json.Marshal([]*common.KeyValue{validCfg})
	require.NoError(t, err)

	validResp := channel.Response{Payload: validCfgBytes}

	t.Run("With --noprompt", func(t *testing.T) {
		c.QueryReturns(validResp, nil)
		c := newMockCmd(t, p, append(args, "--noprompt")...)
		require.NoError(t, c.Execute())
	})

	t.Run("With prompt - Y", func(t *testing.T) {
		c.QueryReturns(validResp, nil)
		w := &mocks.Writer{}
		c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("Y\n")}, w, p, args...)
		require.NoError(t, c.Execute())
		require.Contains(t, w.Written(), msgContinueOrAbort)
		require.Contains(t, w.Written(), msgConfigUpdated)
	})

	t.Run("With prompt - N", func(t *testing.T) {
		c.QueryReturns(validResp, nil)
		w := &mocks.Writer{}
		c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("N\n")}, w, p, args...)
		require.NoError(t, c.Execute())
		require.Contains(t, w.Written(), msgContinueOrAbort)
		require.Contains(t, w.Written(), msgAborted)
		require.NotContains(t, w.Written(), msgConfigUpdated)
	})

	t.Run("With prompt - output stream error", func(t *testing.T) {
		c.QueryReturns(validResp, nil)
		errExpected := errors.New("output stream error")
		w := &mocks.Writer{Err: errExpected}
		c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("N\n")}, w, p, args...)
		require.EqualError(t, c.Execute(), errExpected.Error())
	})

	t.Run("Invalid file index ID", func(t *testing.T) {
		cfg := &common.KeyValue{
			Key:   key,
			Value: &common.Value{TxID: "tx1", Format: "json", Config: mismatchedHandlerCfg},
		}

		cfgBytes, err := json.Marshal([]*common.KeyValue{cfg})
		require.NoError(t, err)

		c.QueryReturns(channel.Response{Payload: cfgBytes}, nil)
		c := newMockCmd(t, p, append(args, "--noprompt")...)

		err = c.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "file index ID must begin with")
	})

	t.Run("File index ID already set", func(t *testing.T) {
		cfg := &common.KeyValue{
			Key:   key,
			Value: &common.Value{TxID: "tx1", Format: "json", Config: setHandlerCfg},
		}

		cfgBytes, err := json.Marshal([]*common.KeyValue{cfg})
		require.NoError(t, err)

		c.QueryReturns(channel.Response{Payload: cfgBytes}, nil)
		c := newMockCmd(t, p, append(args, "--noprompt")...)

		err = c.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "the file index ID for [/content] is already set")
	})
}

func newMockCmd(t *testing.T, p basecmd.FactoryProvider, args ...string) *cobra.Command {
	return newMockCmdWithReaderWriter(t, &mocks.Reader{}, &mocks.Writer{}, p, args...)
}

func newMockCmdWithReaderWriter(t *testing.T, in io.Reader, w io.Writer, p basecmd.FactoryProvider, args ...string) *cobra.Command {
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
