#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

#
# Supported Targets:
#
# build:             builds the binary
# unit-test:         runs unit tests
# integration-test:  runs integration tests
# lint:              runs linters
# checks:            runs build+test+lint
# all :              runs checks+unit-test

GO_CMD ?= go
export GO111MODULE=on

build:
	go build -o bin/hubstore cmd/hub-store/main.go

unit-test:
	go test -count=1 -race -cover -coverprofile=coverage.txt -covermode=atomic ./...

integration-test:
	go test -count=1 -tags=integration ./...

license:
	@scripts/check_license.sh

lint:
	@scripts/check_lint.sh

checks: build license lint

all: checks unit-test
