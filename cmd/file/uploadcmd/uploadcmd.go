/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package uploadcmd

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/hyperledger/fabric-cli/pkg/environment"

	"github.com/trustbloc/sidetree-core-go/pkg/commitment"
	"github.com/trustbloc/sidetree-core-go/pkg/jws"
	"github.com/trustbloc/sidetree-core-go/pkg/patch"
	"github.com/trustbloc/sidetree-core-go/pkg/util/ecsigner"
	"github.com/trustbloc/sidetree-core-go/pkg/util/pubkey"
	"github.com/trustbloc/sidetree-core-go/pkg/versions/0_1/client"

	"github.com/trustbloc/fabric-cli-ext/cmd/basecmd"
	"github.com/trustbloc/fabric-cli-ext/cmd/file/httpclient"
	"github.com/trustbloc/fabric-cli-ext/cmd/file/model"
)

const (
	use      = "upload"
	desc     = "Upload a file to DCAS"
	longDesc = `
The upload command allows a client to upload one or more files to DCAS and add them to a Sidetree file index document. The response is a JSON document that contains the names of the files that were updated along with their DCAS ID and content-type.
`
	examples = `
- Upload two files to the '/content' path and add index entries to the given file index document:
    $ ./fabric file upload --url http://localhost:48326/content --files ./fixtures/testdata/v1/person.schema.json;./fixtures/testdata/v1/raised-hand.png --idxurl http://localhost:48326/file/file:idx:EiAuN66iEpuRt6IIu-2sO3bRM74sS_AIuY6jTbtFUsqAaA== --pwd pwd1 --nextpwd pwd2 --noprompt

	Response:
		[
		  {
			"Name": "person.schema.json",
			"ID": "TbVyraOqG00TacPQH5WwWGnxkszpYSEhBKRyX_f25JI=",
			"ContentType": "application/json"
		  },
		  {
			"Name": "raised-hand.png",
			"ID": "k1fqlkDdtmkTBVTHQgvpJbhTEch2XP0cn0C-DuP-9pE=",
			"ContentType": "image/png"
		  }
		]
`
)

const (
	fileFlag  = "files"
	fileUsage = "The semi-colin separated paths of the files to upload. Example: --files ./samples/content1.json;./samples/image.png"

	urlFlag  = "url"
	urlUsage = "The URL to which to add the file(s). Example: --url http://localhost:48326/content"

	fileIndexURLFlag  = "idxurl"
	fileIndexURLUsage = "The URL of the file index Sidetree document to be updated with the new/updated files. Example: --idxurl http://localhost:48326/file/file:idx:1234"

	authTokenFlag  = "authtoken"
	authTokenUsage = "The bearer authorization token that may be required to access the URL specified by --idxurl. Example: --authtoken mytoken" //nolint: gosec

	contentAuthTokenFlag  = "contentauthtoken"
	contentAuthTokenUsage = "The bearer authorization token to upload files to the URL specified by --url. This is only required if it is different from --authtoken. Example: --contentauthtoken mytoken" //nolint: gosec

	fileIndexNextUpdateKeyFlag  = "nextupdatekey"
	fileIndexNextUpdateKeyUsage = "The public key PEM used for creating commitment for next update of the index document. Example: --nextupdatekey 'MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEFMy2n9jYZChYSjdhK9vUWvPjz9tzBcEa13Ye33haxFsT//3kGxOQhI7yb3MJsDvwLtdfLL6txM3RdOrmLABBvw'"

	fileIndexNextUpdateKeyFileFlag  = "nextupdatekeyfile"
	fileIndexNextUpdateKeyFileUsage = "The file that contains the public key PEM used for creating commitment for next update of the index document. Example: --nextupdatekeyfile ./next_update_public.key"

	fileIndexSigningKeyFlag  = "signingkey"
	fileIndexSigningKeyUsage = "The private key PEM used for signing the update of the index document. Example: --signingkey 'MHcCAQEEILmfa4yss8nsTJK2hKl+LAoiwW3p+eQzaHfITI9z8ptpoAoGCCqGSM49AwEHoUQDQgAEMd1/e/Nxh73bK12PEEcNSY9HxnP0N8er9ww9rjq1tNcsqfRjlL0bdTh9Basfn/4JrQHUHc6uS99yjQc+0u2bVg'"

	fileIndexSigningKeyFileFlag  = "signingkeyfile"
	fileIndexSigningKeyFileUsage = "The file that contains the private key PEM used for signing the update of the index document. Example: --signingkeyfile ./keys/signing.key"

	noPromptFlag  = "noprompt"
	noPromptUsage = "If specified then the upload operation will not prompt for confirmation. Example: --noprompt"

	msgAborted         = "Operation aborted"
	msgContinueOrAbort = "Enter Y to continue or N to abort "

	// default hashes for sidetree
	sha2_256 = 18 // multihash
	sha256   = 5  // hash

	signingAlgorithm = "ES256"

	jsonPatchBasePath  = "/fileIndex/mappings/"
	jsonPatchAddOp     = "add"
	jsonPatchReplaceOp = "replace"
)

var (
	errURLRequired                          = errors.New("URL (--url) is required")
	errFilesRequired                        = errors.New("files (--files) is required")
	errFileIndexURLRequired                 = errors.New("file index URL (--idxurl) is required")
	errNextUpdateKeyOrFileRequired          = errors.New("either next update key (--nextupdatekey) or key file (--nextupdatekeyfile) is required")
	errOnlyOneOfNextUpdateKeyOrFileRequired = errors.New("only one of next update key (--nextupdatekey) or key file (--nextupdatekeyfile) may be specified")
	errNoFileExtension                      = errors.New("content type cannot be deduced since no file extension provided")
	errUnknownExtension                     = errors.New("content type cannot be deduced from extension")
	errSigningKeyOrFileRequired             = errors.New("either signing key (--signingkey) or key file (--signingkeyfile) is required")
	errOnlyOneOfSigningKeyOrFileRequired    = errors.New("only one of signing key (--signingkey) or key file (--signingkeyfile) may be specified")
	errPrivateKeyNotFoundInPEM              = errors.New("private key not found in PEM")
	errPublicKeyNotFoundInPEM               = errors.New("public key not found in PEM")
)

type httpClient interface {
	Post(url string, req []byte, opts ...httpclient.RequestOpt) (*httpclient.HTTPResponse, error)
	Get(url string, opts ...httpclient.RequestOpt) (*httpclient.HTTPResponse, error)
}

// New returns the file upload sub-command
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
			if err := c.validateAndProcessArgs(); err != nil {
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

	cmd.Flags().StringVar(&c.file, fileFlag, "", fileUsage)
	cmd.Flags().StringVar(&c.url, urlFlag, "", urlUsage)
	cmd.Flags().StringVar(&c.authToken, authTokenFlag, "", authTokenUsage)
	cmd.Flags().StringVar(&c.contentAuthToken, contentAuthTokenFlag, "", contentAuthTokenUsage)
	cmd.Flags().StringVar(&c.fileIndexURL, fileIndexURLFlag, "", fileIndexURLUsage)
	cmd.Flags().StringVar(&c.fileIndexNextUpdateKeyString, fileIndexNextUpdateKeyFlag, "", fileIndexNextUpdateKeyUsage)
	cmd.Flags().StringVar(&c.fileIndexNextUpdateKeyFile, fileIndexNextUpdateKeyFileFlag, "", fileIndexNextUpdateKeyFileUsage)
	cmd.Flags().StringVar(&c.fileIndexSigningKeyString, fileIndexSigningKeyFlag, "", fileIndexSigningKeyUsage)
	cmd.Flags().StringVar(&c.fileIndexSigningKeyFile, fileIndexSigningKeyFileFlag, "", fileIndexSigningKeyFileUsage)
	cmd.Flags().BoolVar(&c.noPrompt, noPromptFlag, false, noPromptUsage)

	return cmd
}

// command implements the update command
type command struct {
	*basecmd.Command
	client httpClient

	file                         string
	url                          string
	authToken                    string
	contentAuthToken             string
	basePath                     string
	fileIndexURL                 string
	fileIndexUpdateURL           string
	fileIndexSigningKeyFile      string
	fileIndexSigningKeyString    string
	fileIndexNextUpdateKeyFile   string
	fileIndexNextUpdateKeyString string
	noPrompt                     bool
}

func (c *command) validateAndProcessArgs() error {
	if err := c.validateAndProcessURL(); err != nil {
		return err
	}

	if err := c.validateAndProcessFileIdxURL(); err != nil {
		return err
	}

	if c.file == "" {
		return errFilesRequired
	}

	if err := c.validateSigningKey(); err != nil {
		return err
	}

	if err := c.validateNextUpdateKey(); err != nil {
		return err
	}

	if c.contentAuthToken == "" {
		c.contentAuthToken = c.authToken
	}

	return nil
}

func (c *command) validateAndProcessURL() error {
	if c.url == "" {
		return errURLRequired
	}

	u, err := url.Parse(c.url)
	if err != nil {
		return errors.WithMessagef(err, "invalid URL [%s]", c.url)
	}

	if u.Path == "" {
		return errors.New("invalid URL - no base path found")
	}

	c.basePath = u.Path

	return nil
}

func (c *command) validateAndProcessFileIdxURL() error {
	if c.fileIndexURL == "" {
		return errFileIndexURLRequired
	}

	pos := strings.LastIndex(c.fileIndexURL, "/identifiers")
	if pos == -1 {
		return errors.Errorf("invalid file index URL: [%s] - the file index ID must be prefixed by identifiers/", c.fileIndexURL)
	}

	c.fileIndexUpdateURL = fmt.Sprintf("%s/operations", c.fileIndexURL[0:pos])

	return nil
}

func (c *command) run() error {
	fileIdx, err := c.getFileIndex()
	if err != nil {
		return err
	}

	f, err := c.getFiles()
	if err != nil {
		return err
	}

	if !c.noPrompt {
		confirmed, e := c.confirmUpload(c.url, f)
		if e != nil {
			return e
		}

		if !confirmed {
			return c.Fprintln(msgAborted)
		}
	}

	for _, file := range f {
		id, e := c.upload(file.ContentType, file.Content)
		if e != nil {
			return e
		}

		file.ID = id
	}

	err = c.updateIndexFile(fileIdx, f)
	if err != nil {
		return err
	}

	return c.Fprint(f.String())
}

// confirmUpload prompts the user for confirmation of the upload
func (c *command) confirmUpload(url string, files files) (bool, error) {
	prompt := fmt.Sprintf("Uploading the following files to [%s]\n%s\n%s", url, files, msgContinueOrAbort)

	err := c.Fprintln(prompt)
	if err != nil {
		return false, err
	}

	return strings.ToLower(c.Prompt()) == "y", nil
}

func (c *command) upload(contentType string, fileBytes []byte) (string, error) {
	req := &uploadFile{
		ContentType: contentType,
		Content:     fileBytes,
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	var reqOpts []httpclient.RequestOpt
	if c.contentAuthToken != "" {
		reqOpts = append(reqOpts, httpclient.WithAuthToken(c.contentAuthToken))
	}

	resp, err := c.client.Post(c.url, reqBytes, reqOpts...)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return "", errors.Errorf("status code %d: %s - Did you provide an authorization token (--contentauthtoken)?", resp.StatusCode, resp.ErrorMsg)
		}

		return "", errors.Errorf("status code %d: %s", resp.StatusCode, resp.ErrorMsg)
	}

	var fileID string
	err = json.Unmarshal(resp.Payload, &fileID)
	if err != nil {
		return "", err
	}

	return fileID, nil
}

func (c *command) updateIndexFile(fileIdx *model.FileIndex, files files) error {
	patch, err := getUpdatePatch(fileIdx, files)
	if err != nil {
		return err
	}

	req, err := c.getUpdateRequest(patch)
	if err != nil {
		return err
	}

	var reqOpts []httpclient.RequestOpt
	if c.authToken != "" {
		reqOpts = append(reqOpts, httpclient.WithAuthToken(c.authToken))
	}

	resp, err := c.client.Post(c.fileIndexUpdateURL, req, reqOpts...)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return errors.Errorf("error updating file index document. Status code %d: %s - Did you provide an authorization token (--authtoken)?", resp.StatusCode, resp.ErrorMsg)
		}

		return errors.Errorf("error updating file index document. Status code %d: %s", resp.StatusCode, resp.ErrorMsg)
	}

	return err
}

func (c *command) getFiles() (files, error) {
	var f files
	for _, filePath := range strings.Split(c.file, ";") {
		fileInfo, err := getFileInfo(filePath)
		if err != nil {
			return nil, err
		}

		f = append(f, fileInfo)
	}

	return f, nil
}

func (c *command) getUpdateRequest(patchStr string) ([]byte, error) {
	uniqueSuffix, err := getUniqueSuffix(c.fileIndexURL)
	if err != nil {
		return nil, err
	}

	updatePatch, err := patch.NewJSONPatch(patchStr)
	if err != nil {
		return nil, err
	}

	updateKeySigner, err := c.updateKeySigner()
	if err != nil {
		return nil, err
	}

	updateKeyPublic, err := c.updateKeyPublic()
	if err != nil {
		return nil, err
	}

	nextUpdateKeyPublic, err := c.nextUpdateKeyPublic()
	if err != nil {
		return nil, err
	}

	updateCommitment, err := commitment.Calculate(nextUpdateKeyPublic, sha2_256, sha256)
	if err != nil {
		return nil, err
	}

	return client.NewUpdateRequest(&client.UpdateRequestInfo{
		DidSuffix:        uniqueSuffix,
		UpdateCommitment: updateCommitment,
		UpdateKey:        updateKeyPublic,
		Patches:          []patch.Patch{updatePatch},
		MultihashCode:    sha2_256,
		Signer:           updateKeySigner,
	})
}

func (c *command) getFileIndex() (*model.FileIndex, error) {
	var reqOpts []httpclient.RequestOpt
	if c.authToken != "" {
		reqOpts = append(reqOpts, httpclient.WithAuthToken(c.authToken))
	}

	resp, err := c.client.Get(c.fileIndexURL, reqOpts...)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, errors.Errorf("file index document [%s] not found", c.fileIndexURL)
		}

		if resp.StatusCode == http.StatusUnauthorized {
			return nil, errors.Errorf("error retrieving file index document [%s]. Status code %d: %s - Did you provide an authorization token (--authtoken)?", c.fileIndexURL, resp.StatusCode, resp.ErrorMsg)
		}

		return nil, errors.Errorf("error retrieving file index document [%s] status code %d: %s", c.fileIndexURL, resp.StatusCode, resp.ErrorMsg)
	}

	var r model.DIDResolution
	if errUnmarshal := json.Unmarshal(resp.Payload, &r); errUnmarshal != nil {
		return nil, fmt.Errorf("unmarshal data return from sidtree %w", errUnmarshal)
	}

	didDocBytes := resp.Payload
	// check if data is did resolution
	if len(r.DIDDocument) != 0 {
		didDocBytes = r.DIDDocument
	}

	fileIdxDoc := &model.FileIndexDoc{}
	err = json.Unmarshal(didDocBytes, fileIdxDoc)
	if err != nil {
		return nil, err
	}

	// Validate that the base path is correct
	if fileIdxDoc.FileIndex.BasePath != c.basePath {
		return nil, errors.Errorf("base path of file index doc does not match the base path of the file: [%s] != [%s]", fileIdxDoc.FileIndex.BasePath, c.basePath)
	}

	return &fileIdxDoc.FileIndex, nil
}

func (c *command) updateKeySigner() (client.Signer, error) {
	privateKey, err := c.signingPrivateKey()
	if err != nil {
		return nil, err
	}

	return ecsigner.New(privateKey, signingAlgorithm, model.UpdateKeyID), nil
}

func (c *command) updateKeyPublic() (*jws.JWK, error) {
	privateKey, err := c.signingPrivateKey()
	if err != nil {
		return nil, err
	}

	return pubkey.GetPublicKeyJWK(&privateKey.PublicKey)
}

func (c *command) nextUpdateKeyPublic() (*jws.JWK, error) {
	pubKey, err := c.nextUpdateKey()
	if err != nil {
		return nil, err
	}

	return pubkey.GetPublicKeyJWK(pubKey)
}

func (c *command) nextUpdateKey() (crypto.PublicKey, error) {
	if c.fileIndexNextUpdateKeyFile != "" {
		return publicKeyFromFile(c.fileIndexNextUpdateKeyFile)
	}

	return publicKeyFromPEM([]byte(c.fileIndexNextUpdateKeyString))
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

func (c *command) signingPrivateKey() (*ecdsa.PrivateKey, error) {
	if c.fileIndexSigningKeyFile != "" {
		return privateKeyFromFile(c.fileIndexSigningKeyFile)
	}

	return privateKeyFromPEM([]byte(c.fileIndexSigningKeyString))
}

func (c *command) validateSigningKey() error {
	if c.fileIndexSigningKeyFile == "" && c.fileIndexSigningKeyString == "" {
		return errSigningKeyOrFileRequired
	}

	if c.fileIndexSigningKeyFile != "" && c.fileIndexSigningKeyString != "" {
		return errOnlyOneOfSigningKeyOrFileRequired
	}

	return nil
}

func (c *command) validateNextUpdateKey() error {
	if c.fileIndexNextUpdateKeyFile == "" && c.fileIndexNextUpdateKeyString == "" {
		return errNextUpdateKeyOrFileRequired
	}

	if c.fileIndexNextUpdateKeyFile != "" && c.fileIndexNextUpdateKeyString != "" {
		return errOnlyOneOfNextUpdateKeyOrFileRequired
	}

	return nil
}

func getFileInfo(path string) (*fileInfo, error) {
	var fileName string
	p := strings.LastIndex(path, "/")
	if p == -1 {
		fileName = path
	} else {
		fileName = path[p+1:]
	}

	contentType, err := contentTypeFromFileName(fileName)
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	return &fileInfo{
		Name:        fileName,
		Content:     content,
		ContentType: contentType,
	}, nil
}

func contentTypeFromFileName(fileName string) (string, error) {
	p := strings.LastIndex(fileName, ".")
	if p == -1 {
		return "", errNoFileExtension
	}

	contentType := mime.TypeByExtension(fileName[p:])
	if contentType == "" {
		return "", errUnknownExtension
	}

	return contentType, nil
}

func getUpdatePatch(fileIdx *model.FileIndex, files files) (string, error) {
	var patch []jsonPatch
	for _, f := range files {
		p := jsonPatch{
			Path:  jsonPatchBasePath + f.Name,
			Value: f.ID,
		}

		if _, ok := fileIdx.Mappings[f.Name]; ok {
			p.Op = jsonPatchReplaceOp
		} else {
			p.Op = jsonPatchAddOp
		}

		patch = append(patch, p)
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return "", err
	}

	return string(patchBytes), nil
}

func getUniqueSuffix(id string) (string, error) {
	p := strings.LastIndex(id, ":")
	if p == -1 {
		return "", errors.Errorf("unique suffix not provided in URL [%s]", id)
	}

	return id[p+1:], nil
}

func privateKeyFromFile(file string) (*ecdsa.PrivateKey, error) {
	keyBytes, err := ioutil.ReadFile(filepath.Clean(file))
	if err != nil {
		return nil, err
	}

	return privateKeyFromPEM(keyBytes)
}

func privateKeyFromPEM(privateKeyPEM []byte) (*ecdsa.PrivateKey, error) {
	privBlock, _ := pem.Decode(privateKeyPEM)
	if privBlock == nil {
		return nil, errPrivateKeyNotFoundInPEM
	}

	privKey, err := x509.ParseECPrivateKey(privBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return privKey, nil
}
