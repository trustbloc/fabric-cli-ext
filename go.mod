// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/fabric-cli-ext

require (
	github.com/hyperledger/fabric-cli v0.0.0-20191215205855-97c039341083
	github.com/hyperledger/fabric-sdk-go v1.0.0-beta1.0.20200222173625-ff3bdd738791
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.4
	github.com/stretchr/testify v1.4.0
	github.com/trustbloc/sidetree-core-go v0.1.3-0.20200318151633-ccbb84596468
)

go 1.13

replace github.com/hyperledger/fabric-protos-go => github.com/trustbloc/fabric-protos-go-ext v0.1.2

replace github.com/sirupsen/logrus => github.com/sirupsen/logrus v1.3.0

replace golang.org/x/crypto => golang.org/x/crypto v0.0.0-20190617133340-57b3e21c3d56

replace golang.org/x/sys => golang.org/x/sys v0.0.0-20190616124812-15dcb6c0061f

replace gopkg.in/check.v1 => gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127

replace github.com/stretchr/testify => github.com/stretchr/testify v1.4.0
