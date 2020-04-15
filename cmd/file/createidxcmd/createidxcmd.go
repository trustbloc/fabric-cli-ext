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

	"github.com/hyperledger/fabric-cli/pkg/environment"

	"github.com/trustbloc/sidetree-core-go/pkg/document"
	"github.com/trustbloc/sidetree-core-go/pkg/jws"
	"github.com/trustbloc/sidetree-core-go/pkg/restapi/helper"
	"github.com/trustbloc/sidetree-core-go/pkg/util/pubkey"

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

	recoveryPWDFlag  = "recoverypwd"
	recoveryPWDUsage = "The password for recovery of the document. Example: --recoverypwd myrecoverypwd"

	recoveryKeyFlag  = "recoverykey"
	recoveryKeyUsage = "The public key PEM used for recovery of the document. Example: --recoverykey 'MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEXlp4fWF5rgLthKr20tsJ0tBIE6UmrGuAC8iVG/DaedkSt7HihCx/t2BGjooduaKwEIOmPjx2zBsbkbFrYhhnVw'"

	recoveryKeyFileFlag  = "recoverykeyfile"
	recoveryKeyFileUsage = "The file that contains the public key PEM used for recovery of the document. Example: --recoverykeyfile ./recovery_public.key"

	nextUpdatePWDFlag  = "nextpwd"
	nextUpdatePWDUsage = "The password for the next update of the document. Example: --nextpwd pwd2"

	updateKeyFlag  = "updatekey"
	updateKeyUsage = "The public key PEM used for validating the signature of the next update of the document. Example: --updatekey 'MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEFMy2n9jYZChYSjdhK9vUWvPjz9tzBcEa13Ye33haxFsT//3kGxOQhI7yb3MJsDvwLtdfLL6txM3RdOrmLABBvw'"

	updateKeyFileFlag  = "updatekeyfile"
	updateKeyFileUsage = "The file that contains the public key PEM used for validating the signature of the next update of the document. Example: --updatekeyfile ./update_public.key"

	noPromptFlag  = "noprompt"
	noPromptUsage = "If specified then the operation will not prompt for confirmation. Example: --noprompt"

	msgAborted         = "Operation aborted"
	msgContinueOrAbort = "Enter Y to continue or N to abort "

	sha2_256 = 18

	publicKeyField    = "publicKey"
	publicKeyTemplate = `[{"id":"%s","type":"JwsVerificationKey2020","usage":["ops"],"publicKeyJwk":%s}]`
)

var (
	errURLRequired                        = errors.New("URL (--url) is required")
	errRecoveryPWDRequired                = errors.New("recovery password (--recoverypwd) is required")
	errNextUpdatePWDRequired              = errors.New("next update password (--nextpwd) is required")
	errPathRequired                       = errors.New("path (--path) is required")
	errInvalidPath                        = errors.New("path (--path) must begin with '/'")
	errRecoveryKeyOrFileRequired          = errors.New("either recovery key (--recoverykey) or key file (--recoverykeyfile) is required")
	errOnlyOneOfRecoveryKeyOrFileRequired = errors.New("only one of recovery key (--recoverykey) or key file (--recoverykeyfile) may be specified")
	errUpdateKeyOrFileRequired            = errors.New("either update key (--updatekey) or key file (--updatekeyfile) is required")
	errOnlyOneOfUpdateKeyOrFileRequired   = errors.New("only one of Update key (--updatekey) or key file (--updatekeyfile) may be specified")
	errPublicKeyNotFoundInPEM             = errors.New("public key not found in PEM")
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
	cmd.Flags().StringVar(&c.recoveryPWD, recoveryPWDFlag, "", recoveryPWDUsage)
	cmd.Flags().StringVar(&c.recoveryKeyString, recoveryKeyFlag, "", recoveryKeyUsage)
	cmd.Flags().StringVar(&c.recoveryKeyFile, recoveryKeyFileFlag, "", recoveryKeyFileUsage)
	cmd.Flags().StringVar(&c.nextUpdatePWD, nextUpdatePWDFlag, "", nextUpdatePWDUsage)
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
	recoveryPWD       string
	nextUpdatePWD     string
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

	if c.recoveryPWD == "" {
		return errRecoveryPWDRequired
	}

	if c.nextUpdatePWD == "" {
		return errNextUpdatePWDRequired
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
	doc, err := c.getOpaqueDocument(content)
	if err != nil {
		return nil, err
	}

	recoveryKey, err := c.recoveryKeyJWK()
	if err != nil {
		return nil, err
	}

	return helper.NewCreateRequest(
		&helper.CreateRequestInfo{
			OpaqueDocument:          doc,
			RecoveryKey:             recoveryKey,
			NextRecoveryRevealValue: []byte(c.recoveryPWD),
			NextUpdateRevealValue:   []byte(c.nextUpdatePWD),
			MultihashCode:           sha2_256,
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
	updatePublicKey, err := c.updateKeyJWK()
	if err != nil {
		return "", err
	}

	publicKeyBytes, err := json.Marshal(updatePublicKey)
	if err != nil {
		return "", err
	}

	publicKeysStr := fmt.Sprintf(publicKeyTemplate, model.UpdateKeyID, string(publicKeyBytes))

	var publicKeys []map[string]interface{}
	err = json.Unmarshal([]byte(publicKeysStr), &publicKeys)
	if err != nil {
		return "", err
	}

	doc, err := document.FromBytes([]byte(content))
	if err != nil {
		return "", err
	}

	doc[publicKeyField] = publicKeys

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
