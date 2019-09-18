/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-cli/pkg/environment"
	"github.com/hyperledger/fabric-cli/pkg/fabric"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-cli-ext/cmd/ledgerconfig/mocks"
)

func TestCriteriaBaseCommand(t *testing.T) {
	t.Run("No options", func(t *testing.T) {
		require.EqualError(t, newMockCriteriaCmd(t, &testCmd{}, nil).Execute(), errMspOrCriteriaRequired)
	})

	t.Run("--mspid with --criteria", func(t *testing.T) {
		require.EqualError(t, newMockCriteriaCmd(t, &testCmd{}, nil, "--mspid", "MSP1", "--criteria", "{}").Execute(), errCriteriaMustBeAlone)
	})

	t.Run("--peerid with --criteria", func(t *testing.T) {
		require.EqualError(t, newMockCriteriaCmd(t, &testCmd{}, nil, "--peerid", "peer1", "--criteria", "{}").Execute(), errCriteriaMustBeAlone)
	})

	t.Run("--appname with --criteria", func(t *testing.T) {
		require.EqualError(t, newMockCriteriaCmd(t, &testCmd{}, nil, "--appname", "app1", "--criteria", "{}").Execute(), errCriteriaMustBeAlone)
	})

	t.Run("--appver with --criteria", func(t *testing.T) {
		require.EqualError(t, newMockCriteriaCmd(t, &testCmd{}, nil, "--appver", "v1", "--criteria", "{}").Execute(), errCriteriaMustBeAlone)
	})

	t.Run("--componentname with --criteria", func(t *testing.T) {
		require.EqualError(t, newMockCriteriaCmd(t, &testCmd{}, nil, "--componentname", "comp1", "--criteria", "{}").Execute(), errCriteriaMustBeAlone)
	})

	t.Run("--componentver with --criteria", func(t *testing.T) {
		require.EqualError(t, newMockCriteriaCmd(t, &testCmd{}, nil, "--componentver", "v1", "--criteria", "{}").Execute(), errCriteriaMustBeAlone)
	})

	t.Run("Invalid --criteria", func(t *testing.T) {
		c := newMockCriteriaCmd(t, &testCmd{}, nil, "--criteria", "xxx")
		err := c.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), errInvalidCriteria)
	})

	t.Run("GetCriteriaBytes with --criteria => valid", func(t *testing.T) {
		const expectedBytes = `{"MspID":"MSP1","PeerID":"peer1","AppName":"app1","AppVersion":"v1","ComponentName":"comp1","ComponentVersion":"v1"}`
		mc := &testCmd{}
		c := newMockCriteriaCmd(t, mc, nil, "--criteria", expectedBytes)
		err := c.Execute()
		require.NoError(t, err)

		bytes, err := mc.GetCriteriaBytes()
		require.NoError(t, err)
		require.Equal(t, expectedBytes, string(bytes))
	})

	t.Run("GetCriteriaBytes with flags => valid", func(t *testing.T) {
		const msp = "MSP1"
		const peer = "peer1"
		const app = "app1"
		const version = "v1"
		const comp = "comp1"

		expectedBytes := fmt.Sprintf(`{"MspID":"%s","PeerID":"%s","AppName":"%s","AppVersion":"%s","ComponentName":"%s","ComponentVersion":"%s"}`, msp, peer, app, version, comp, version)
		mc := &testCmd{}
		c := newMockCriteriaCmd(t, mc, nil, "--mspid", msp, "--peerid", peer, "--appname", app, "--appver", version, "--componentname", comp, "--componentver", version)
		err := c.Execute()
		require.NoError(t, err)

		bytes, err := mc.GetCriteriaBytes()
		require.NoError(t, err)
		require.Equal(t, expectedBytes, string(bytes))
	})
}

func TestCriteriaBaseCommand_GetConfig(t *testing.T) {
	factory := &mocks.Factory{}
	p := func(config *environment.Config) (fabric.Factory, error) { return factory, nil }

	t.Run("GetConfig => valid", func(t *testing.T) {
		const payload = `[{"MspID":"msp1"}]`
		ch := &mocks.Channel{}
		resp := channel.Response{Payload: []byte(payload)}
		ch.QueryReturns(resp, nil)
		factory.ChannelReturns(ch, nil)

		mc := &testCmd{}
		c := newMockCriteriaCmd(t, mc, p)
		require.NotNil(t, c)
		bytes, err := mc.GetConfig([]byte(`["MspID":"msp1"}`))
		require.NoError(t, err)
		require.Equal(t, payload, string(bytes))
	})

	t.Run("GetConfig => channel error", func(t *testing.T) {
		errExpected := errors.New("channel error")
		factory.ChannelReturns(nil, errExpected)

		mc := &testCmd{}
		c := newMockCriteriaCmd(t, mc, p)
		require.NotNil(t, c)
		_, err := mc.GetConfig([]byte(`["MspID":"msp1"}`))
		require.EqualError(t, err, errExpected.Error())
	})

	t.Run("GetConfig => query error", func(t *testing.T) {
		errExpected := errors.New("query error")
		ch := &mocks.Channel{}
		ch.QueryReturns(channel.Response{}, errExpected)
		factory.ChannelReturns(ch, nil)

		mc := &testCmd{}
		c := newMockCriteriaCmd(t, mc, p)
		require.NotNil(t, c)
		_, err := mc.GetConfig([]byte(`["MspID":"msp1"}`))
		require.EqualError(t, err, errExpected.Error())
	})
}

type testCmd struct {
	*CriteriaBaseCommand
}

func newMockCriteriaCmd(t *testing.T, c *testCmd, p FactoryProvider, args ...string) *cobra.Command {
	cmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.Validate()
		},
	}
	c.CriteriaBaseCommand = newMockCriteriaBaseCmd(t, cmd, p, args...)
	return cmd
}

func newMockCriteriaBaseCmd(t *testing.T, cmd *cobra.Command, p FactoryProvider, args ...string) *CriteriaBaseCommand {
	settings := environment.NewDefaultSettings()
	settings.Streams.In = &mocks.Reader{}
	settings.Streams.Out = &mocks.Writer{}

	settings.Config.CurrentContext = "testctx"
	settings.Config.Contexts[settings.Config.CurrentContext] = &environment.Context{}

	cmd.SetArgs(args)

	c := NewCriteriaBaseCommand(settings, p, cmd)
	require.NotNil(t, c)

	return c
}
