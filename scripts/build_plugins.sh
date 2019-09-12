#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -e

echo "Building plugins..."

# ledgerconfig
go build -buildmode=plugin -o ./.build/ledgerconfig/ledgerconfig.so ./cmd/ledgerconfig/ledgerconfig.go
cp ./cmd/ledgerconfig/plugin.yaml ./.build/ledgerconfig/
