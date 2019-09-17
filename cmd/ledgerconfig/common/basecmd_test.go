/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"io"
	"testing"

	"github.com/hyperledger/fabric-cli/pkg/environment"
	"github.com/hyperledger/fabric-cli/pkg/fabric"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-cli-ext/cmd/ledgerconfig/mocks"
)

//go:generate counterfeiter -o ../mocks/channel.gen.go --fake-name Channel github.com/hyperledger/fabric-cli/pkg/fabric.Channel
//go:generate counterfeiter -o ../mocks/factory.gen.go --fake-name Factory github.com/hyperledger/fabric-cli/pkg/fabric.Factory

func TestBaseCommand_Channel(t *testing.T) {
	t.Run("With factory error", func(t *testing.T) {
		errExpected := errors.New("factory error")
		p := func(config *environment.Config) (fabric.Factory, error) { return nil, errExpected }
		c := newMockCmd(t, &mocks.Writer{}, p)
		ch, err := c.Channel()
		require.EqualError(t, err, errExpected.Error())
		require.Nil(t, ch)
	})
	t.Run("With channel error", func(t *testing.T) {
		errExpected := errors.New("channel error")
		factory := &mocks.Factory{}
		factory.ChannelReturns(nil, errExpected)

		p := func(config *environment.Config) (fabric.Factory, error) { return factory, nil }
		c := newMockCmd(t, &mocks.Writer{}, p)
		ch, err := c.Channel()
		require.EqualError(t, err, errExpected.Error())
		require.Nil(t, ch)
	})
	t.Run("Channel", func(t *testing.T) {
		factory := &mocks.Factory{}
		factory.ChannelReturns(&mocks.Channel{}, nil)

		p := func(config *environment.Config) (fabric.Factory, error) { return factory, nil }
		c := newMockCmd(t, &mocks.Writer{}, p)
		ch, err := c.Channel()
		require.NoError(t, err)
		require.NotNil(t, ch)
	})
}

func TestBaseCommand_Context(t *testing.T) {
	p := func(config *environment.Config) (fabric.Factory, error) { return &mocks.Factory{}, nil }
	c := newMockCmd(t, &mocks.Writer{}, p)
	ctx := c.Context()
	require.NotNil(t, ctx)
}

func TestBaseCommand_Fprintln(t *testing.T) {
	const msg = "written message"

	w := &mocks.Writer{}
	p := func(config *environment.Config) (fabric.Factory, error) { return &mocks.Factory{}, nil }
	c := newMockCmd(t, w, p)
	require.NoError(t, c.Fprintln(msg))
	require.Equal(t, msg, w.Written())
}

func newMockCmd(t *testing.T, out io.Writer, p FactoryProvider) *BaseCommand {
	settings := environment.NewDefaultSettings()
	settings.Streams.Out = out

	settings.Config.CurrentContext = "testctx"
	settings.Config.Contexts[settings.Config.CurrentContext] = &environment.Context{}

	c := NewBaseCmd(settings, p)
	require.NotNil(t, c)

	return c
}
