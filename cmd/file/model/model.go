/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package model

import "encoding/json"

// FileIndexDoc contains a file index document
type FileIndexDoc struct {
	ID           string    `json:"id"`
	UniqueSuffix string    `json:"did_suffix"`
	FileIndex    FileIndex `json:"fileIndex"`
}

// FileIndex contains the mappings of file name to ID
type FileIndex struct {
	BasePath string            `json:"basePath"`
	Mappings map[string]string `json:"mappings,omitempty"`
}

// DIDResolution did resolution
type DIDResolution struct {
	Context          interface{}     `json:"@context"`
	DIDDocument      json.RawMessage `json:"didDocument"`
	ResolverMetadata json.RawMessage `json:"resolverMetadata"`
	MethodMetadata   json.RawMessage `json:"methodMetadata"`
}
