// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/fabric-cli-ext

require (
	github.com/hyperledger/fabric-cli v0.0.0-20190920195049-94768c835ab2
	github.com/hyperledger/fabric-protos-go v0.0.0-20190823190507-26c33c998676 // indirect
	github.com/hyperledger/fabric-sdk-go v1.0.0-beta1.0.20190930220855-cea2ffaf627c
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.4
	github.com/stretchr/testify v1.3.0
)

go 1.13

replace github.com/hyperledger/fabric-protos-go => github.com/trustbloc/fabric-protos-go-ext v0.0.0-20191001172134-1815f5c382ff
