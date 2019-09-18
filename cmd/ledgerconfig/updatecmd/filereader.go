/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package updatecmd

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func readFile(path string) ([]byte, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, errors.WithMessagef(err, "error opening file [%s]", path)
	}
	defer func() {
		if e := file.Close(); e != nil {
			// This shouldn't happen
			panic(err.Error())
		}
	}()

	configBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.WithMessagef(err, "error reading config file [%s]", path)
	}
	return configBytes, nil
}
