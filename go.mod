// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/fabric-cli-ext

require (
	github.com/hyperledger/fabric-cli v0.0.0-20191215205855-97c039341083
	github.com/hyperledger/fabric-sdk-go v1.0.0-beta1.0.20200222173625-ff3bdd738791
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.4
	github.com/stretchr/testify v1.3.0
)

go 1.13

replace github.com/hyperledger/fabric-protos-go => github.com/trustbloc/fabric-protos-go-ext v0.1.2
