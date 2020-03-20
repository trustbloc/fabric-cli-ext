/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package httpclient

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

// HTTPResponse contains an HTTP response
type HTTPResponse struct {
	StatusCode  int
	Payload     []byte
	ErrorMsg    string
	ContentType string
}

// Client is an HTTP client
type Client struct {
	client *http.Client
}

// Opt defines an option for the HTTP client
type Opt func(c *Client)

// WithTransport sets the transport for the client. This is usually only set for unit tests.
func WithTransport(rt http.RoundTripper) Opt {
	return func(c *Client) {
		c.client.Transport = rt
	}
}

// New returns a new HTTP client
func New(opts ...Opt) *Client {
	c := &Client{
		client: &http.Client{},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Post posts an HTTP request
func (c *Client) Post(url string, req []byte) (*HTTPResponse, error) {
	resp, err := c.send(url, req)
	if err != nil {
		return nil, err
	}
	return c.handle(resp)
}

// Get send an HTTP GET request
func (c *Client) Get(url string) (*HTTPResponse, error) {
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	return c.handle(resp)
}

func (c *Client) handle(resp *http.Response) (*HTTPResponse, error) {
	gotBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body failed: %s", err)
	}

	if status := resp.StatusCode; status != http.StatusOK {
		return &HTTPResponse{
			StatusCode: status,
			ErrorMsg:   string(gotBody),
		}, nil
	}

	var contentType string
	hVal := resp.Header["Content-Type"]
	if len(hVal) > 0 {
		contentType = hVal[0]
	}

	return &HTTPResponse{
		StatusCode:  http.StatusOK,
		Payload:     gotBody,
		ContentType: contentType,
	}, nil
}

func (c *Client) send(url string, req []byte) (*http.Response, error) {
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(req))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	return c.client.Do(httpReq)
}
