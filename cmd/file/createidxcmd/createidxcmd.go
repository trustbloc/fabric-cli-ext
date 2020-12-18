/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package createidxcmd

import (
	"crypto"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/trustbloc/sidetree-core-go/pkg/commitment"
	"github.com/trustbloc/sidetree-core-go/pkg/document"
	"github.com/trustbloc/sidetree-core-go/pkg/jws"
	"github.com/trustbloc/sidetree-core-go/pkg/util/pubkey"
	"github.com/trustbloc/sidetree-core-go/pkg/versions/0_1/client"

	"github.com/hyperledger/fabric-cli/pkg/environment"
	"github.com/trustbloc/fabric-cli-ext/cmd/basecmd"
	"github.com/trustbloc/fabric-cli-ext/cmd/file/httpclient"
	"github.com/trustbloc/fabric-cli-ext/cmd/file/model"
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
		  "id": "file:idx:EiAuN66iEpuRt6IIu-2sO3bRM74sS_AIuY6jTbtFUsqAaA=="
		}
`
)

const (
	urlFlag  = "url"
	urlUsage = "The URL of the file index Sidetree endpoint. Example: --url http://localhost:48326/file"

	pathFlag  = "path"
	pathUsage = "The base path of the endpoint that will be indexed by this document. Example: --path /schema"

	authTokenFlag  = "authtoken"
	authTokenUsage = "The bearer authorization token that may be required to access some HTTP endpoints. Example: --authtoken mytoken" //nolint: gosec

	recoveryKeyFlag  = "recoverykey"
	recoveryKeyUsage = "The public key PEM used for recovery of the document. Example: --recoverykey 'MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEXlp4fWF5rgLthKr20tsJ0tBIE6UmrGuAC8iVG/DaedkSt7HihCx/t2BGjooduaKwEIOmPjx2zBsbkbFrYhhnVw'"

	recoveryKeyFileFlag  = "recoverykeyfile"
	recoveryKeyFileUsage = "The file that contains the public key PEM used for recovery of the document. Example: --recoverykeyfile ./recovery_public.key"

	updateKeyFlag  = "updatekey"
	updateKeyUsage = "The public key PEM used for validating the signature of the next update of the document. Example: --updatekey 'MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEFMy2n9jYZChYSjdhK9vUWvPjz9tzBcEa13Ye33haxFsT//3kGxOQhI7yb3MJsDvwLtdfLL6txM3RdOrmLABBvw'"

	updateKeyFileFlag  = "updatekeyfile"
	updateKeyFileUsage = "The file that contains the public key PEM used for validating the signature of the next update of the document. Example: --updatekeyfile ./update_public.key"

	noPromptFlag  = "noprompt"
	noPromptUsage = "If specified then the operation will not prompt for confirmation. Example: --noprompt"

	msgAborted         = "Operation aborted"
	msgContinueOrAbort = "Enter Y to continue or N to abort "

	sha2_256 = 18
)

var (
	errURLRequired                        = errors.New("URL (--url) is required")
	errPathRequired                       = errors.New("path (--path) is required")
	errInvalidPath                        = errors.New("path (--path) must begin with '/'")
	errRecoveryKeyOrFileRequired          = errors.New("either recovery key (--recoverykey) or key file (--recoverykeyfile) is required")
	errOnlyOneOfRecoveryKeyOrFileRequired = errors.New("only one of recovery key (--recoverykey) or key file (--recoverykeyfile) may be specified")
	errUpdateKeyOrFileRequired            = errors.New("either update key (--updatekey) or key file (--updatekeyfile) is required")
	errOnlyOneOfUpdateKeyOrFileRequired   = errors.New("only one of update key (--updatekey) or key file (--updatekeyfile) may be specified")
	errPublicKeyNotFoundInPEM             = errors.New("public key not found in PEM")
)

type httpClient interface {
	Post(url string, req []byte, opts ...httpclient.RequestOpt) (*httpclient.HTTPResponse, error)
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
	cmd.Flags().StringVar(&c.authToken, authTokenFlag, "", authTokenUsage)
	cmd.Flags().StringVar(&c.recoveryKeyString, recoveryKeyFlag, "", recoveryKeyUsage)
	cmd.Flags().StringVar(&c.recoveryKeyFile, recoveryKeyFileFlag, "", recoveryKeyFileUsage)
	cmd.Flags().StringVar(&c.updateKeyString, updateKeyFlag, "", updateKeyUsage)
	cmd.Flags().StringVar(&c.updateKeyFile, updateKeyFileFlag, "", updateKeyFileUsage)
	cmd.Flags().BoolVar(&c.noPrompt, noPromptFlag, false, noPromptUsage)

	return cmd
}

// command implements the update command
type command struct {
	*basecmd.Command
	client httpClient

	// Flags
	url               string
	path              string
	authToken         string
	noPrompt          bool
	recoveryKeyFile   string
	recoveryKeyString string
	updateKeyFile     string
	updateKeyString   string
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

	if err := c.validateRecoveryKey(); err != nil {
		return err
	}

	if err := c.validateUpdateKey(); err != nil {
		return err
	}

	return nil
}

func (c *command) run() error {
	fileIdxDoc := &model.FileIndexDoc{
		FileIndex: model.FileIndex{
			BasePath: c.path,
			Mappings: map[string]string{
				".": c.path,
			},
		},
	}

	docBytes, err := json.Marshal(fileIdxDoc)
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

	resp, err := c.post(req)
	if err != nil {
		return err
	}

	didDocBytes, err := c.getDoc(resp.Payload)
	if err != nil {
		return err
	}

	if err := c.Fprint(string(didDocBytes)); err != nil {
		return err
	}

	return nil
}

func (c *command) post(data []byte) (*httpclient.HTTPResponse, error) {
	var reqOpts []httpclient.RequestOpt
	if c.authToken != "" {
		reqOpts = append(reqOpts, httpclient.WithAuthToken(c.authToken))
	}

	resp, err := c.client.Post(c.url, data, reqOpts...)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, errors.Errorf("status code %d: %s - Did you provide an authorization token (--authtoken)?", resp.StatusCode, resp.ErrorMsg)
		}

		return nil, errors.Errorf("status code %d: %s", resp.StatusCode, resp.ErrorMsg)
	}

	return resp, nil
}

func (c *command) getDoc(payload []byte) ([]byte, error) {
	var r model.DIDResolution
	if errUnmarshal := json.Unmarshal(payload, &r); errUnmarshal != nil {
		return nil, fmt.Errorf("unmarshal data return from sidtree %w", errUnmarshal)
	}

	didDocBytes := payload
	// check if data is did resolution
	if len(r.DIDDocument) != 0 {
		didDocBytes = r.DIDDocument
	}

	return didDocBytes, nil
}

func (c *command) newCreateRequest(content string) ([]byte, error) {
	doc, err := c.getOpaqueDocument(content)
	if err != nil {
		return nil, err
	}

	recoveryKey, err := c.recoveryKeyJWK()
	if err != nil {
		return nil, err
	}

	recoveryCommitment, err := commitment.GetCommitment(recoveryKey, sha2_256)
	if err != nil {
		return nil, err
	}

	updateKey, err := c.updateKeyJWK()
	if err != nil {
		return nil, err
	}

	updateCommitment, err := commitment.GetCommitment(updateKey, sha2_256)
	if err != nil {
		return nil, err
	}

	return client.NewCreateRequest(
		&client.CreateRequestInfo{
			OpaqueDocument:     doc,
			RecoveryCommitment: recoveryCommitment,
			UpdateCommitment:   updateCommitment,
			MultihashCode:      sha2_256,
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

func (c *command) recoveryKeyJWK() (*jws.JWK, error) {
	publicKey, err := c.recoveryPublicKey()
	if err != nil {
		return nil, err
	}

	return pubkey.GetPublicKeyJWK(publicKey)
}

func (c *command) recoveryPublicKey() (crypto.PublicKey, error) {
	if c.recoveryKeyFile != "" {
		return publicKeyFromFile(c.recoveryKeyFile)
	}

	return publicKeyFromPEM([]byte(c.recoveryKeyString))
}

func (c *command) updateKeyJWK() (*jws.JWK, error) {
	publicKey, err := c.updatePublicKey()
	if err != nil {
		return nil, err
	}

	return pubkey.GetPublicKeyJWK(publicKey)
}

func (c *command) updatePublicKey() (crypto.PublicKey, error) {
	if c.updateKeyFile != "" {
		return publicKeyFromFile(c.updateKeyFile)
	}

	return publicKeyFromPEM([]byte(c.updateKeyString))
}

func (c *command) getOpaqueDocument(content string) (string, error) {
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

func (c *command) validateUpdateKey() error {
	if c.updateKeyFile == "" && c.updateKeyString == "" {
		return errUpdateKeyOrFileRequired
	}

	if c.updateKeyFile != "" && c.updateKeyString != "" {
		return errOnlyOneOfUpdateKeyOrFileRequired
	}

	return nil
}

func (c *command) validateRecoveryKey() error {
	if c.recoveryKeyFile == "" && c.recoveryKeyString == "" {
		return errRecoveryKeyOrFileRequired
	}

	if c.recoveryKeyFile != "" && c.recoveryKeyString != "" {
		return errOnlyOneOfRecoveryKeyOrFileRequired
	}

	return nil
}

func publicKeyFromFile(file string) (crypto.PublicKey, error) {
	keyBytes, err := ioutil.ReadFile(filepath.Clean(file))
	if err != nil {
		return nil, err
	}

	return publicKeyFromPEM(keyBytes)
}

func publicKeyFromPEM(pubKeyPEM []byte) (crypto.PublicKey, error) {
	block, _ := pem.Decode(pubKeyPEM)
	if block == nil {
		return nil, errPublicKeyNotFoundInPEM
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	publicKey, ok := key.(crypto.PublicKey)
	if !ok {
		return nil, errors.New("invalid public key")
	}

	return publicKey, nil
}
