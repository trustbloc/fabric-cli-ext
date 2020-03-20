/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package uploadcmd

import (
	"encoding/json"
)

type uploadFile struct {
	ContentType string `json:"contentType"`
	Content     []byte `json:"content"`
}

type fileInfo struct {
	Name        string `json:",omitempty"`
	ID          string `json:",omitempty"`
	ContentType string `json:",omitempty"`
	Content     []byte `json:",omitempty"`
}

type jsonPatch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

type files []*fileInfo

func (f files) String() string {
	ff := make([]*fileInfo, len(f))
	for i, info := range f {
		ff[i] = &fileInfo{
			Name:        info.Name,
			ID:          info.ID,
			ContentType: info.ContentType,
		}
	}

	bytes, err := json.Marshal(ff)
	if err != nil {
		panic(err)
	}

	return string(bytes)
}
