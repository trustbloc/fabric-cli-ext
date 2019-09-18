#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

#
# Supported Targets:
#
# all:                        runs checks, unit tests, and builds the plugins
# plugins:                    builds fabric-cli and plugins
# unit-test:                  runs unit tests
# lint:                       runs linters
# checks:                     runs code checks
# generate:                   generates mocks
#

# Local variables used by makefile
PROJECT_NAME            = fabric-cli-ext
export GO111MODULE      = on
export FABRIC_CLI_VERSION ?= f6d60d55e800403c587b564c1ca383b2cb496bed

checks: version license lint

lint:
	@scripts/check_lint.sh

license: version
	@scripts/check_license.sh

all: clean checks unit-test plugins

unit-test:
	@scripts/unit.sh

version:
	@scripts/check_version.sh

plugins:
	@scripts/build_plugins.sh

clean:
	rm -rf ./.build

generate:
	go generate ./...

.PHONY: all version unit-test license plugins clean generate
