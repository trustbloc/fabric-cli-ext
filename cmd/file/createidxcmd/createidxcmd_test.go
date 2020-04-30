/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package createidxcmd

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/fabric-cli/pkg/environment"

	"github.com/trustbloc/fabric-cli-ext/cmd/file/httpclient"
	"github.com/trustbloc/fabric-cli-ext/cmd/file/model"
	"github.com/trustbloc/fabric-cli-ext/cmd/mocks"
)

const (
	recoveryPublicKey = `
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEbENaETENCgl8+qgls5JBgogX8Vp1
G8qXPRBB6W9pzfiphvbPl52B9PLZAWFLcHsP3jsdhag9KNSeVKrQtRshPw==
-----END PUBLIC KEY-----`

	updatePublicKey = `
-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEbT4kKzrPhR/YFWdHHjxtRHUdsOlt
gCw04H3xwMXlHY8fIQwQdrKXsNrG482lIFu2tVkKoj51EGiMZUP7jcqp0w==
-----END PUBLIC KEY-----`
)

func TestCreateIDXCmd_New(t *testing.T) {
	require.NotNil(t, New(environment.NewDefaultSettings()))
}

func TestCreateIDXCmd_InvalidOptions(t *testing.T) {
	const (
		urlFlag             = "--url"
		url                 = "http://localhost:80/file"
		pathFlag            = "--path"
		path                = "/content"
		recoverypwdFlag     = "--recoverypwd"
		nextupdatepwdFlag   = "--nextpwd"
		pwd                 = "pwd1"
		recoverykeyFlag     = "--recoverykey"
		recoverykeyfileFlag = "--recoverykeyfile"
		updatekeyFlag       = "--updatekey"
		updatekeyfileFlag   = "--updatekeyfile"
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
		require.EqualError(t, newMockCmd(t, nil, urlFlag, url, pathFlag, path).Execute(), errRecoveryPWDRequired.Error())
	})

	t.Run("Next update password required", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, nil, urlFlag, url, pathFlag, path, recoverypwdFlag, pwd).Execute(), errNextUpdatePWDRequired.Error())
	})

	t.Run("Recovery key required", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, nil, urlFlag, url, pathFlag, path, recoverypwdFlag, pwd, nextupdatepwdFlag, pwd).Execute(), errRecoveryKeyOrFileRequired.Error())
	})

	t.Run("Recovery key and file specified", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, nil, urlFlag, url, pathFlag, path, recoverypwdFlag, pwd, nextupdatepwdFlag, pwd, recoverykeyFlag, recoveryPublicKey, recoverykeyfileFlag, "./key").Execute(), errOnlyOneOfRecoveryKeyOrFileRequired.Error())
	})

	t.Run("Update key required", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, nil, urlFlag, url, pathFlag, path, recoverypwdFlag, pwd, nextupdatepwdFlag, pwd, recoverykeyFlag, recoveryPublicKey).Execute(), errUpdateKeyOrFileRequired.Error())
	})

	t.Run("Update key and file specified", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, nil, urlFlag, url, pathFlag, path, recoverypwdFlag, pwd, nextupdatepwdFlag, pwd, recoverykeyFlag, recoveryPublicKey, updatekeyFlag, updatePublicKey, updatekeyfileFlag, "./key").Execute(), errOnlyOneOfUpdateKeyOrFileRequired.Error())
	})
}

func TestCreateIDXCmd(t *testing.T) {
	fileIdxDoc := &model.FileIndexDoc{
		ID:           "file:idx:EiAuN66iEpuRt6IIu-2sO3bRM74sS_AIuY6jTbtFUsqAaA==",
		UniqueSuffix: "EiAuN66iEpuRt6IIu-2sO3bRM74sS_AIuY6jTbtFUsqAaA==",
		FileIndex:    model.FileIndex{BasePath: "/content"},
	}

	fileIndexBytes, err := json.Marshal(fileIdxDoc)
	require.NoError(t, err)

	didResolution := model.DIDResolution{DIDDocument: fileIndexBytes}

	didResolutionBytes, err := json.Marshal(didResolution)
	require.NoError(t, err)

	args := []string{"--url", "http://localhost:80/file", "--path", "/content", "--recoverypwd", "pwd1", "--nextpwd", "pwd1", "--recoverykey", recoveryPublicKey, "--updatekey", updatePublicKey, "--authtoken", "mytoken"}
	header := map[string][]string{"Content-Type": {"application/json"}}

	transport := mocks.NewTransport().WithPostResponse(
		&http.Response{
			StatusCode: http.StatusOK,
			Header:     header,
			Body:       mocks.NewResponseBody(didResolutionBytes),
		},
	)

	t.Run("With prompt - Y", func(t *testing.T) {
		w := &mocks.Writer{}

		c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("Y\n")}, w, transport, args...)
		require.NoError(t, c.Execute())
		require.Contains(t, w.Written(), msgContinueOrAbort)
		require.Contains(t, w.Written(), string(fileIndexBytes))
	})

	t.Run("With --noprompt", func(t *testing.T) {
		w := &mocks.Writer{}

		c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("Y\n")}, w, transport, append(args, "--noprompt")...)
		require.NoError(t, c.Execute())
		require.NotContains(t, w.Written(), msgContinueOrAbort)
		require.Contains(t, w.Written(), string(fileIndexBytes))
	})

	t.Run("With invalid key", func(t *testing.T) {
		w := &mocks.Writer{}

		args := []string{"--url", "http://localhost:80/file", "--path", "/content", "--recoverypwd", "pwd1", "--nextpwd", "pwd1", "--recoverykey", recoveryPublicKey, "--updatekey", "xxx"}
		c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("Y\n")}, w, transport, args...)
		err := c.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), errPublicKeyNotFoundInPEM.Error())
	})

	t.Run("With key files", func(t *testing.T) {
		args := []string{"--url", "http://localhost:80/file", "--path", "/content", "--recoverypwd", "pwd1", "--nextpwd", "pwd1", "--noprompt"}

		t.Run("Update key file not found -> error", func(t *testing.T) {
			w := &mocks.Writer{}
			c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("Y\n")}, w, transport, append(args, "--recoverykeyfile", "../testdata/recover_public.key", "--updatekeyfile", "../testdata/xxx.key")...)
			err := c.Execute()
			require.Error(t, err)
			require.Contains(t, err.Error(), "no such file or directory")
		})

		t.Run("Recovery key file not found -> error", func(t *testing.T) {
			w := &mocks.Writer{}
			c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("Y\n")}, w, transport, append(args, "--recoverykeyfile", "../testdata/xxx.key", "--updatekeyfile", "../testdata/update_public.key")...)
			err := c.Execute()
			require.Error(t, err)
			require.Contains(t, err.Error(), "no such file or directory")
			require.Contains(t, w.Written(), err.Error())
		})

		t.Run("Success", func(t *testing.T) {
			w := &mocks.Writer{}
			c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("Y\n")}, w, transport, append(args, "--recoverykeyfile", "../testdata/recover_public.key", "--updatekeyfile", "../testdata/update_public.key")...)
			require.NoError(t, c.Execute())
			require.NotContains(t, w.Written(), msgContinueOrAbort)
			require.Contains(t, w.Written(), string(fileIndexBytes))
		})
	})

	t.Run("With prompt - N", func(t *testing.T) {
		w := &mocks.Writer{}

		c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("N\n")}, w, nil, args...)
		require.NoError(t, c.Execute())
		require.Contains(t, w.Written(), msgContinueOrAbort)
		require.Contains(t, w.Written(), msgAborted)
		require.NotContains(t, w.Written(), fileIndexBytes)
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
