#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -e

echo "Building fabric-cli..."

mkdir -p .build/bin
cd .build
rm -rf fabric-cli
git clone https://github.com/hyperledger/fabric-cli.git
cd fabric-cli
git checkout $FABRIC_CLI_VERSION
sed  -e "\$areplace github.com/hyperledger/fabric-sdk-go => github.com/hyperledger/fabric-sdk-go v1.0.0-beta1.0.20191219180315-e1055f391525" -i go.mod
sed  -e "\$areplace github.com/hyperledger/fabric-protos-go => github.com/trustbloc/fabric-protos-go-ext v0.1.1" -i go.mod
make
cp ./bin/fabric ../bin/fabric
cd ../

echo "Building plugins..."

# The plugin needs to import exactly the same source code as was used to build fabric-cli, otherwise
# an error will result when the plugin is loaded. So, copy fabric-cli-ext and modify go.mod to
# replace github/hyperledger/fabric-cli with the local copy.
mkdir ./fabric-cli-ext
cp -r ../cmd/ ./fabric-cli-ext/cmd/
cp ../go.mod ./fabric-cli-ext/
cd ./fabric-cli-ext
sed  -e "\$areplace github.com/hyperledger/fabric-cli => ..\/fabric-cli" -i go.mod

# ledgerconfig
go build -buildmode=plugin -o ../ledgerconfig/ledgerconfig.so ./cmd/ledgerconfig/ledgerconfig.go
cp ./cmd/ledgerconfig/plugin.yaml ../ledgerconfig/

cd ..
rm -rf ./fabric-cli
rm -rf ./fabric-cli-ext
