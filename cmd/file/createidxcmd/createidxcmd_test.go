/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package createidxcmd

import (
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/fabric-cli/pkg/environment"

	"github.com/trustbloc/fabric-cli-ext/cmd/file/httpclient"
	"github.com/trustbloc/fabric-cli-ext/cmd/mocks"
)

func TestCreateIDXCmd_New(t *testing.T) {
	require.NotNil(t, New(environment.NewDefaultSettings()))
}

func TestCreateIDXCmd_InvalidOptions(t *testing.T) {
	const (
		urlFlag         = "--url"
		url             = "http://localhost:80/file"
		pathFlag        = "--path"
		path            = "/content"
		recoverypwdFlag = "--recoverypwd"
		pwd             = "pwd1"
	)

	t.Run("No options", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, nil).Execute(), errURLRequired.Error())
	})

	t.Run("No path", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, nil, urlFlag, url).Execute(), errPathRequired.Error())
	})

	t.Run("Invalid path", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, nil, urlFlag, url, pathFlag, "content").Execute(), errInvalidPath.Error())
	})

	t.Run("Recovery password required", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, nil, urlFlag, url, pathFlag, path).Execute(), errRecoveryOTPRequired.Error())
	})

	t.Run("Next update password required", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, nil, urlFlag, url, pathFlag, path, recoverypwdFlag, pwd).Execute(), errNextUpdateOTPRequired.Error())
	})
}

func TestCreateIDXCmd(t *testing.T) {
	const doc = `{".":"/content","id":"file:idx:EiAuN66iEpuRt6IIu-2sO3bRM74sS_AIuY6jTbtFUsqAaA==","published":false}`

	args := []string{"--url", "http://localhost:80/file", "--path", "/content", "--recoverypwd", "pwd1", "--nextpwd", "pwd1"}
	header := map[string][]string{"Content-Type": {"application/json"}}

	t.Run("With prompt - Y", func(t *testing.T) {
		transport := mocks.NewTransport().WithPostResponse(
			&http.Response{
				StatusCode: http.StatusOK,
				Header:     header,
				Body:       mocks.NewResponseBody([]byte(doc)),
			},
		)

		w := &mocks.Writer{}

		c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("Y\n")}, w, transport, args...)
		require.NoError(t, c.Execute())
		require.Contains(t, w.Written(), msgContinueOrAbort)
		require.Contains(t, w.Written(), doc)
	})

	t.Run("With --noprompt", func(t *testing.T) {
		transport := mocks.NewTransport().WithPostResponse(
			&http.Response{
				StatusCode: http.StatusOK,
				Header:     header,
				Body:       mocks.NewResponseBody([]byte(doc)),
			},
		)

		w := &mocks.Writer{}

		c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("Y\n")}, w, transport, append(args, "--noprompt")...)
		require.NoError(t, c.Execute())
		require.NotContains(t, w.Written(), msgContinueOrAbort)
		require.Contains(t, w.Written(), doc)
	})

	t.Run("With prompt - N", func(t *testing.T) {
		w := &mocks.Writer{}

		c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("N\n")}, w, nil, args...)
		require.NoError(t, c.Execute())
		require.Contains(t, w.Written(), msgContinueOrAbort)
		require.Contains(t, w.Written(), msgAborted)
		require.NotContains(t, w.Written(), doc)
	})

	t.Run("With prompt - output stream error", func(t *testing.T) {
		errExpected := errors.New("output stream error")
		w := &mocks.Writer{Err: errExpected}
		c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("N\n")}, w, nil, args...)
		require.EqualError(t, c.Execute(), errExpected.Error())
	})

	t.Run("With client error", func(t *testing.T) {
		errExpected := errors.New("injected error")

		transport := mocks.NewTransport().WithPostError(errExpected)

		c := newMockCmd(t, transport, append(args, "--noprompt")...)
		err := c.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
	})

	t.Run("With HTTP error", func(t *testing.T) {
		expectedResponse := "server error"

		transport := mocks.NewTransport().WithPostResponse(
			&http.Response{
				StatusCode: http.StatusInternalServerError,
				Header:     header,
				Body:       mocks.NewResponseBody([]byte(expectedResponse)),
			},
		)

		c := newMockCmd(t, transport, append(args, "--noprompt")...)
		err := c.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), expectedResponse)
	})
}

func newMockCmd(t *testing.T, rt http.RoundTripper, args ...string) *cobra.Command {
	return newMockCmdWithReaderWriter(t, &mocks.Reader{}, &mocks.Writer{}, rt, args...)
}

func newMockCmdWithReaderWriter(t *testing.T, in io.Reader, w io.Writer, transport http.RoundTripper, args ...string) *cobra.Command {
	settings := environment.NewDefaultSettings()
	settings.Streams.Out = w
	settings.Streams.In = in

	settings.Config.CurrentContext = "testctx"
	settings.Config.Contexts[settings.Config.CurrentContext] = &environment.Context{}

	c := newCmd(settings, httpclient.New(httpclient.WithTransport(transport)))
	require.NotNil(t, c)

	c.SetArgs(args)

	return c
}
