/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package uploadcmd

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

func TestUloadCmd_New(t *testing.T) {
	require.NotNil(t, New(environment.NewDefaultSettings()))
}

func TestUloadCmd_InvalidOptions(t *testing.T) {
	const (
		urlFlag    = "--url"
		url        = "http://localhost:80/content"
		filesFlag  = "--files"
		files      = "./samplefile.json"
		idxUrlFlag = "--idxurl"
		idxUrl     = "http://localhost:80/file/file:idx:1234"
		pwdFlag    = "--pwd"
		pwd        = "pwd1"
	)

	t.Run("No options", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, nil).Execute(), errURLRequired.Error())
	})

	t.Run("Invalid --url", func(t *testing.T) {
		err := newMockCmd(t, nil, urlFlag, "localhost:80").Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid URL")
	})

	t.Run("No --files", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, nil, urlFlag, url).Execute(), errFilesRequired.Error())
	})

	t.Run("No --idxurl", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, nil, urlFlag, url, filesFlag, files).Execute(), errFileIndexURLRequired.Error())
	})

	t.Run("Invalid --idxurl", func(t *testing.T) {
		err := newMockCmd(t, nil, urlFlag, url, filesFlag, files, idxUrlFlag, "localhost:80").Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid file index URL")
	})

	t.Run("No --pwd", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, nil, urlFlag, url, filesFlag, files, idxUrlFlag, idxUrl).Execute(), errFileIndexUpdateOTPRequired.Error())
	})

	t.Run("No --nextpwd", func(t *testing.T) {
		require.EqualError(t, newMockCmd(t, nil, urlFlag, url, filesFlag, files, idxUrlFlag, idxUrl, pwdFlag, pwd).Execute(), errFileIndexNextUpdateOTPRequired.Error())
	})
}

func TestUploadCmd(t *testing.T) {
	const (
		url        = "http://localhost:48326/content/v1"
		files      = "./testdata/person.schema.json"
		idxUrl     = "http://localhost:48326/file/file:idx:EiAuN66iEpuRt6IIu-2sO3bRM74sS_AIuY6jTbtFUsqAaA=="
		resp       = `[{"Name":"person.schema.json","ID":"TbVyraOqG00TacPQH5WwWGnxkszpYSEhBKRyX_f25JI=","ContentType":"application/json"}]`
		dcasIDJSON = `"TbVyraOqG00TacPQH5WwWGnxkszpYSEhBKRyX_f25JI="`
	)

	var (
		args   = []string{"--url", url, "--files", files, "--idxurl", idxUrl, "--pwd", "pwd1", "--nextpwd", "pwd2"}
		header = map[string][]string{"Content-Type": {"application/json"}}
	)

	fileIdxDoc := &model.FileIndexDoc{
		ID:           "file:idx:EiAuN66iEpuRt6IIu-2sO3bRM74sS_AIuY6jTbtFUsqAaA==",
		UniqueSuffix: "EiAuN66iEpuRt6IIu-2sO3bRM74sS_AIuY6jTbtFUsqAaA==",
		FileIndex:    model.FileIndex{BasePath: "/content/v1"},
	}

	fileIdxDocBytes, err := json.Marshal(fileIdxDoc)
	require.NoError(t, err)

	transport := mocks.NewTransport().
		WithGetResponse(
			&http.Response{
				StatusCode: http.StatusOK,
				Header:     header,
				Body:       mocks.NewResponseBody(fileIdxDocBytes),
			},
		).
		WithPostResponse(
			&http.Response{
				StatusCode: http.StatusOK,
				Header:     header,
				Body:       mocks.NewResponseBody([]byte(dcasIDJSON)),
			},
		)

	t.Run("With prompt - Y", func(t *testing.T) {
		w := &mocks.Writer{}

		c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("Y\n")}, w, transport, args...)
		require.NoError(t, c.Execute())
		require.Contains(t, w.Written(), msgContinueOrAbort)
		require.Contains(t, w.Written(), resp)
	})

	t.Run("With --noprompt", func(t *testing.T) {
		w := &mocks.Writer{}

		c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("Y\n")}, w, transport, append(args, "--noprompt")...)
		require.NoError(t, c.Execute())
		require.NotContains(t, w.Written(), msgContinueOrAbort)
		require.Contains(t, w.Written(), resp)
	})

	t.Run("With prompt - N", func(t *testing.T) {
		w := &mocks.Writer{}

		c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("N\n")}, w, transport, args...)
		require.NoError(t, c.Execute())
		require.Contains(t, w.Written(), msgContinueOrAbort)
		require.Contains(t, w.Written(), msgAborted)
		require.NotContains(t, w.Written(), resp)
	})

	t.Run("With prompt - output stream error", func(t *testing.T) {
		errExpected := errors.New("output stream error")
		w := &mocks.Writer{Err: errExpected}
		c := newMockCmdWithReaderWriter(t, &mocks.Reader{Bytes: []byte("N\n")}, w, transport, args...)
		require.EqualError(t, c.Execute(), errExpected.Error())
	})

	t.Run("With GET error", func(t *testing.T) {
		errExpected := errors.New("injected error")

		transport := mocks.NewTransport().WithGetError(errExpected)

		c := newMockCmd(t, transport, append(args, "--noprompt")...)
		err := c.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
	})

	t.Run("With POST error", func(t *testing.T) {
		errExpected := errors.New("injected error")

		transport := mocks.NewTransport().
			WithGetResponse(
				&http.Response{
					StatusCode: http.StatusOK,
					Header:     header,
					Body:       mocks.NewResponseBody(fileIdxDocBytes),
				},
			).
			WithPostError(errExpected)

		c := newMockCmd(t, transport, append(args, "--noprompt")...)
		err := c.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
	})

	t.Run("With HTTP error", func(t *testing.T) {
		expectedResponse := "server error"

		transport := mocks.NewTransport().WithGetResponse(
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

	t.Run("File IDX doc not found", func(t *testing.T) {
		transport := mocks.NewTransport().WithGetResponse(
			&http.Response{
				StatusCode: http.StatusNotFound,
				Header:     header,
				Body:       mocks.NewResponseBody([]byte("not found")),
			},
		)

		c := newMockCmd(t, transport, append(args, "--noprompt")...)
		err := c.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "not found")
	})

	t.Run("With invalid base path", func(t *testing.T) {
		fileIdxDoc := &model.FileIndexDoc{
			ID:           "file:idx:EiAuN66iEpuRt6IIu-2sO3bRM74sS_AIuY6jTbtFUsqAaA==",
			UniqueSuffix: "EiAuN66iEpuRt6IIu-2sO3bRM74sS_AIuY6jTbtFUsqAaA==",
			FileIndex: model.FileIndex{
				BasePath: "/schema",
			},
		}

		mismatchedFileIDXDoc, err := json.Marshal(fileIdxDoc)
		require.NoError(t, err)

		transport := mocks.NewTransport().
			WithGetResponse(
				&http.Response{
					StatusCode: http.StatusOK,
					Header:     header,
					Body:       mocks.NewResponseBody(mismatchedFileIDXDoc),
				},
			)

		c := newMockCmd(t, transport, append(args, "--noprompt")...)
		err = c.Execute()
		require.Error(t, err)
		require.Contains(t, err.Error(), "base path of file index doc does not match the base path of the file")
	})
}

func TestContentTypeFromFileName(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		contentType, err := contentTypeFromFileName("file.json")
		require.NoError(t, err)
		require.Equal(t, "application/json", contentType)
	})

	t.Run("No extension -> error", func(t *testing.T) {
		contentType, err := contentTypeFromFileName("file")
		require.EqualError(t, err, errNoFileExtension.Error())
		require.Empty(t, contentType)
	})

	t.Run("No extension -> error", func(t *testing.T) {
		contentType, err := contentTypeFromFileName("file.xxx")
		require.EqualError(t, err, errUnknownExtension.Error())
		require.Empty(t, contentType)
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
