/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"io"
	"net/http"
)

// MockTransport implements a mock HTTP transport
type MockTransport struct {
	GetResponse  *http.Response
	GetErr       error
	PostResponse *http.Response
	PostErr      error
}

// NewTransport returns a mock transport
func NewTransport() *MockTransport {
	return &MockTransport{}
}

// WithGetResponse sets the mock response for a Get
func (m *MockTransport) WithGetResponse(resp *http.Response) *MockTransport {
	m.GetResponse = resp
	return m
}

// WithPostResponse sets the mock response for a Post
func (m *MockTransport) WithPostResponse(resp *http.Response) *MockTransport {
	m.PostResponse = resp
	return m
}

// WithGetError injects an error
func (m *MockTransport) WithGetError(err error) *MockTransport {
	m.GetErr = err
	return m
}

// WithPostError injects an error
func (m *MockTransport) WithPostError(err error) *MockTransport {
	m.PostErr = err
	return m
}

// RoundTrip implements http.RoundTripper
func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method == http.MethodPost {
		return m.PostResponse, m.PostErr
	}

	return m.GetResponse, m.GetErr
}

// MockResponseBody implements a mock io.ReadCloser
type MockResponseBody struct {
	Bytes []byte
	Err   error
}

// NewResponseBody returns a new mock response body
func NewResponseBody(content []byte) *MockResponseBody {
	return &MockResponseBody{
		Bytes: content,
	}
}

// Read reads from the byte array
func (m *MockResponseBody) Read(p []byte) (int, error) {
	if m.Err != nil {
		return 0, m.Err
	}

	return copy(p, m.Bytes), io.EOF
}

// Close mocks out the Close func
func (m *MockResponseBody) Close() error {
	return m.Err
}
