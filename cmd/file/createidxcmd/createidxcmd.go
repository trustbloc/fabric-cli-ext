/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package createidxcmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hyperledger/fabric-cli/pkg/environment"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/trustbloc/sidetree-core-go/pkg/document"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/helper"

	"github.com/trustbloc/fabric-cli-ext/cmd/basecmd"
	"github.com/trustbloc/fabric-cli-ext/cmd/file/httpclient"
)

const (
	use      = "createidx"
	desc     = "Creates a file index document in Sidetree and returns the document"
	longDesc = `
The create command allows a client to create a new file index document in Sidetree. A single entry is added to the file index document whose name is '.' and value is the specified base path. This is done to ensure uniqueness of the initial document and for validation once file mappings are added.
`
	examples = `
- Create a file index document:
    $ ./fabric-cli file createidx --path /content --url http://localhost:48326/file --recoverypwd pwd1 --nextpwd pwd1 --noprompt

	Response:
		{
		  ".": "/content",
		  "id": "file:idx:EiAuN66iEpuRt6IIu-2sO3bRM74sS_AIuY6jTbtFUsqAaA==",
		  "published": false
		}
`
)

const (
	urlFlag  = "url"
	urlUsage = "The URL of the file index Sidetree endpoint. Example: --url http://localhost:48326/file"

	pathFlag  = "path"
	pathUsage = "The base path of the endpoint that will be indexed by this document. Example: --path /schema"

	recoveryOTPFlag  = "recoverypwd"
	recoveryOTPUsage = "The password for recovery of the document. Example: --recoverypwd myrecoverypwd"

	nextUpdateOTPFlag  = "nextpwd"
	nextUpdateOTPUsage = "The password for the next update of the document. Example: --nextpwd pwd2"

	noPromptFlag  = "noprompt"
	noPromptUsage = "If specified then the operation will not prompt for confirmation. Example: --noprompt"

	msgAborted         = "Operation aborted"
	msgContinueOrAbort = "Enter Y to continue or N to abort "

	sha2_256 = 18
)

var (
	errURLRequired           = errors.New("URL (--url) is required")
	errRecoveryOTPRequired   = errors.New("recovery password (--recoverypwd) is required")
	errNextUpdateOTPRequired = errors.New("next update password (--nextpwd) is required")
	errPathRequired          = errors.New("path (--path) is required")
	errInvalidPath           = errors.New("path (--path) must begin with '/'")
)

type httpClient interface {
	Post(url string, req []byte) (*httpclient.HTTPResponse, error)
}

// New returns the file createidx sub-command
func New(settings *environment.Settings) *cobra.Command {
	return newCmd(settings, httpclient.New())
}

func newCmd(settings *environment.Settings, client httpClient) *cobra.Command {
	c := &command{
		Command: basecmd.New(settings, nil),
		client:  client,
	}

	cmd := &cobra.Command{
		Use:     use,
		Short:   desc,
		Long:    longDesc,
		Example: examples,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := c.validate(); err != nil {
				return err
			}
			return nil
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return c.run()
		},
	}

	c.Settings = settings
	cmd.SetOutput(c.Settings.Streams.Out)
	cmd.SilenceUsage = true

	cmd.Flags().StringVar(&c.url, urlFlag, "", urlUsage)
	cmd.Flags().StringVar(&c.path, pathFlag, "", pathUsage)
	cmd.Flags().StringVar(&c.recoveryOTP, recoveryOTPFlag, "", recoveryOTPUsage)
	cmd.Flags().StringVar(&c.nextUpdateOTP, nextUpdateOTPFlag, "", nextUpdateOTPUsage)
	cmd.Flags().BoolVar(&c.noPrompt, noPromptFlag, false, noPromptUsage)

	return cmd
}

// command implements the update command
type command struct {
	*basecmd.Command
	client httpClient

	// Flags
	url           string
	path          string
	recoveryOTP   string
	nextUpdateOTP string
	noPrompt      bool
}

func (c *command) validate() error {
	if c.url == "" {
		return errURLRequired
	}

	if c.path == "" {
		return errPathRequired
	}

	if c.path[0:1] != "/" {
		return errInvalidPath
	}

	if c.recoveryOTP == "" {
		return errRecoveryOTPRequired
	}

	if c.nextUpdateOTP == "" {
		return errNextUpdateOTPRequired
	}

	return nil
}

func (c *command) run() error {
	doc := make(map[string]interface{})
	// Add one mapping which is the path where this index will be used
	doc["."] = c.path

	docBytes, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	req, err := c.newCreateRequest(string(docBytes))
	if err != nil {
		return err
	}

	if !c.noPrompt {
		confirmed, e := c.confirm()
		if e != nil {
			return e
		}

		if !confirmed {
			return c.Fprintln(msgAborted)
		}
	}

	resp, err := c.client.Post(c.url, req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("status code %d: %s", resp.StatusCode, resp.ErrorMsg)
	}

	if err := c.Fprint(string(resp.Payload)); err != nil {
		return err
	}

	return nil
}

func (c *command) newCreateRequest(content string) ([]byte, error) {
	doc, err := getOpaqueDocument(content)
	if err != nil {
		return nil, err
	}

	return helper.NewCreateRequest(
		&helper.CreateRequestInfo{
			OpaqueDocument:  doc,
			RecoveryKey:     "recoveryKey", // Should this be hard-coded?
			NextRecoveryOTP: base64.URLEncoding.EncodeToString([]byte(c.recoveryOTP)),
			MultihashCode:   sha2_256,
			NextUpdateOTP:   base64.URLEncoding.EncodeToString([]byte(c.nextUpdateOTP)),
		},
	)
}
func (c *command) confirm() (bool, error) {
	prompt := fmt.Sprintf("Creating file index document for path [%s]\n%s", c.path, msgContinueOrAbort)

	err := c.Fprintln(prompt)
	if err != nil {
		return false, err
	}

	return strings.ToLower(c.Prompt()) == "y", nil
}

func getOpaqueDocument(content string) (string, error) {
	doc, err := document.FromBytes([]byte(content))

	if err != nil {
		return "", err
	}

	bytes, err := doc.Bytes()
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
