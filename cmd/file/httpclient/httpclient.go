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

const (
	authHeader  = "Authorization"
	tokenPrefix = "Bearer "
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

type requestOptions struct {
	authToken string
}

// RequestOpt sets a request option
type RequestOpt func(opts *requestOptions)

// WithAuthToken sets an authorization token in the header
func WithAuthToken(token string) RequestOpt {
	return func(opts *requestOptions) {
		opts.authToken = token
	}
}

// Post posts an HTTP request
func (c *Client) Post(url string, req []byte, opts ...RequestOpt) (*HTTPResponse, error) {
	resp, err := c.put(url, req, opts)
	if err != nil {
		return nil, err
	}
	return c.handle(resp)
}

// Get put an HTTP GET request
func (c *Client) Get(url string, opts ...RequestOpt) (*HTTPResponse, error) {
	resp, err := c.get(url, opts)
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

func (c *Client) get(url string, opts []RequestOpt) (*http.Response, error) {
	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	options := resolveRequestOptions(opts)

	if options.authToken != "" {
		httpReq.Header.Set(authHeader, tokenPrefix+options.authToken)
	}

	return c.client.Do(httpReq)
}

func (c *Client) put(url string, req []byte, opts []RequestOpt) (*http.Response, error) {
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(req))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	options := resolveRequestOptions(opts)

	if options.authToken != "" {
		httpReq.Header.Set(authHeader, tokenPrefix+options.authToken)
	}

	return c.client.Do(httpReq)
}

func resolveRequestOptions(opts []RequestOpt) *requestOptions {
	options := &requestOptions{}

	for _, opt := range opts {
		opt(options)
	}

	return options
}
