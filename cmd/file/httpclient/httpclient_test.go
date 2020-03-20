/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package httpclient

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/fabric-cli-ext/cmd/mocks"
)

func TestClient(t *testing.T) {
	reqData := []byte("some request")
	respData := []byte("some response")

	header := map[string][]string{"Content-Type": {"application/json"}}

	t.Run("Get -> success", func(t *testing.T) {
		transport := mocks.NewTransport().
			WithGetResponse(
				&http.Response{
					StatusCode: http.StatusOK,
					Header:     header,
					Body:       mocks.NewResponseBody(respData),
				},
			)

		c := New(WithTransport(transport))
		require.NotNil(t, c)

		resp, err := c.Get("http://localhost:80")
		require.NoError(t, err)
		require.Equal(t, respData, resp.Payload)
	})

	t.Run("Post -> success", func(t *testing.T) {
		transport := mocks.NewTransport().
			WithPostResponse(
				&http.Response{
					StatusCode: http.StatusOK,
					Header:     header,
					Body:       mocks.NewResponseBody(respData),
				},
			)

		c := New(WithTransport(transport))
		require.NotNil(t, c)

		resp, err := c.Post("http://localhost:80", reqData)
		require.NoError(t, err)
		require.Equal(t, respData, resp.Payload)
	})

	t.Run("Get error code", func(t *testing.T) {
		const errMessage = "some error"
		transport := mocks.NewTransport().
			WithGetResponse(
				&http.Response{
					StatusCode: http.StatusInternalServerError,
					Header:     header,
					Body:       mocks.NewResponseBody([]byte(errMessage)),
				},
			)

		c := New(WithTransport(transport))
		require.NotNil(t, c)

		resp, err := c.Get("http://localhost:80")
		require.NoError(t, err)
		require.Nil(t, resp.Payload)
		require.Equal(t, errMessage, resp.ErrorMsg)
	})

	t.Run("Get read error -> error", func(t *testing.T) {
		errExpected := errors.New("injected error")

		transport := mocks.NewTransport().
			WithGetResponse(
				&http.Response{
					StatusCode: http.StatusOK,
					Header:     header,
					Body:       &mocks.MockResponseBody{Err: errExpected},
				},
			)

		c := New(WithTransport(transport))
		require.NotNil(t, c)

		_, err := c.Get("http://localhost:80")
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
	})
}
