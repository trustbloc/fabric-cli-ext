/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package basecmd

import (
	"io"
	"testing"

	"github.com/hyperledger/fabric-cli/pkg/environment"
	"github.com/hyperledger/fabric-cli/pkg/fabric"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-cli-ext/cmd/mocks"
)

//go:generate counterfeiter -o ../mocks/channel.gen.go --fake-name Channel github.com/hyperledger/fabric-cli/pkg/fabric.Channel
//go:generate counterfeiter -o ../mocks/factory.gen.go --fake-name Factory github.com/hyperledger/fabric-cli/pkg/fabric.Factory

func TestBaseCommand_Channel(t *testing.T) {
	t.Run("With factory error", func(t *testing.T) {
		errExpected := errors.New("factory error")
		p := func(config *environment.Config) (fabric.Factory, error) { return nil, errExpected }
		c := newMockCmd(t, p)
		ch, err := c.Channel()
		require.EqualError(t, err, errExpected.Error())
		require.Nil(t, ch)
	})
	t.Run("With channel error", func(t *testing.T) {
		errExpected := errors.New("channel error")
		factory := &mocks.Factory{}
		factory.ChannelReturns(nil, errExpected)

		p := func(config *environment.Config) (fabric.Factory, error) { return factory, nil }
		c := newMockCmd(t, p)
		ch, err := c.Channel()
		require.EqualError(t, err, errExpected.Error())
		require.Nil(t, ch)
	})
	t.Run("Channel", func(t *testing.T) {
		factory := &mocks.Factory{}
		factory.ChannelReturns(&mocks.Channel{}, nil)

		p := func(config *environment.Config) (fabric.Factory, error) { return factory, nil }
		c := newMockCmd(t, p)
		ch, err := c.Channel()
		require.NoError(t, err)
		require.NotNil(t, ch)
	})
}

func TestBaseCommand_Context(t *testing.T) {
	p := func(config *environment.Config) (fabric.Factory, error) { return &mocks.Factory{}, nil }
	c := newMockCmd(t, p)
	ctx := c.Context()
	require.NotNil(t, ctx)
}

func TestBaseCommand_Fprintln(t *testing.T) {
	const msg = "written message"

	p := func(config *environment.Config) (fabric.Factory, error) { return &mocks.Factory{}, nil }

	t.Run("No error", func(t *testing.T) {
		w := &mocks.Writer{}
		c := newMockCmdWithReaderWriter(t, &mocks.Reader{}, w, p)
		require.NoError(t, c.Fprintln(msg))
		require.Equal(t, msg, w.Written())
	})

	t.Run("With error", func(t *testing.T) {
		errExpected := errors.New("write error")
		w := &mocks.Writer{Err: errExpected}
		c := newMockCmdWithReaderWriter(t, &mocks.Reader{}, w, p)
		require.EqualError(t, c.Fprintln(msg), errExpected.Error())
	})
}

func TestBaseCommand_FprintlnOrPanic(t *testing.T) {
	const msg = "written message"

	p := func(config *environment.Config) (fabric.Factory, error) { return &mocks.Factory{}, nil }

	t.Run("No panic", func(t *testing.T) {
		w := &mocks.Writer{}
		c := newMockCmdWithReaderWriter(t, &mocks.Reader{}, w, p)
		require.NotPanics(t, func() {
			c.FprintlnOrPanic(msg)
		})
		require.Equal(t, msg, w.Written())
	})

	t.Run("With panic", func(t *testing.T) {
		errExpected := errors.New("write error")
		w := &mocks.Writer{Err: errExpected}
		c := newMockCmdWithReaderWriter(t, &mocks.Reader{}, w, p)
		require.PanicsWithValue(t, errExpected.Error(), func() {
			c.FprintlnOrPanic(msg)
		})
	})
}

func TestBaseCommand_Fprint(t *testing.T) {
	const msg = "written message"

	p := func(config *environment.Config) (fabric.Factory, error) { return &mocks.Factory{}, nil }

	t.Run("No error", func(t *testing.T) {
		w := &mocks.Writer{}
		c := newMockCmdWithReaderWriter(t, &mocks.Reader{}, w, p)
		require.NoError(t, c.Fprint(msg))
		require.Equal(t, msg, w.Written())
	})

	t.Run("With error", func(t *testing.T) {
		errExpected := errors.New("write error")
		w := &mocks.Writer{Err: errExpected}
		c := newMockCmdWithReaderWriter(t, &mocks.Reader{}, w, p)
		require.EqualError(t, c.Fprint(msg), errExpected.Error())
	})
}

func TestBaseCommand_FprintOrPanic(t *testing.T) {
	const msg = "written message"

	p := func(config *environment.Config) (fabric.Factory, error) { return &mocks.Factory{}, nil }

	t.Run("No panic", func(t *testing.T) {
		w := &mocks.Writer{}
		c := newMockCmdWithReaderWriter(t, &mocks.Reader{}, w, p)
		require.NotPanics(t, func() {
			c.FprintOrPanic(msg)
		})
		require.Equal(t, msg, w.Written())
	})

	t.Run("With panic", func(t *testing.T) {
		errExpected := errors.New("write error")
		w := &mocks.Writer{Err: errExpected}
		c := newMockCmdWithReaderWriter(t, &mocks.Reader{}, w, p)
		require.PanicsWithValue(t, errExpected.Error(), func() {
			c.FprintOrPanic(msg)
		})
	})
}

func TestBaseCommand_Prompt(t *testing.T) {
	p := func(config *environment.Config) (fabric.Factory, error) { return &mocks.Factory{}, nil }
	c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("Y\n")}, &mocks.Writer{}, p)
	require.Equal(t, "Y", c.Prompt())
}

func newMockCmd(t *testing.T, p FactoryProvider) *Command {
	return newMockCmdWithReaderWriter(t, &mocks.Reader{}, &mocks.Writer{}, p)
}

func newMockCmdWithReaderWriter(t *testing.T, in io.Reader, out io.Writer, p FactoryProvider) *Command {
	settings := environment.NewDefaultSettings()
	settings.Streams.In = in
	settings.Streams.Out = out

	settings.Config.CurrentContext = "testctx"
	settings.Config.Contexts[settings.Config.CurrentContext] = &environment.Context{}

	c := New(settings, p)
	require.NotNil(t, c)

	return c
}
