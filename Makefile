#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

#
# Supported Targets:
#
# build:              builds the binary
# unit-test:          runs unit tests
# integration-test:   runs integration tests
# lint:               runs linters
# checks:             runs build+test+lint
# generate-test-keys: generates keys for testing
# clean:              removes test fixtures
# all :               runs checks+unit-test


GO_CMD ?= go
export GO111MODULE=on


#couchdb image parameters
#this couchdb image contains startup scripts that autorun and create all necessary db artifacts
export HUB_STORE_COUCHDB_IMAGE ?= trustbloc/hub-store-couchdb

export PKGS=`go list github.com/trustbloc/hub-store/... `

build:
	go build -o bin/hubstore cmd/hub-store/main.go

//TODO : Separate the couchdb test (which are dependant on external dependencies ) as integration test
unit-test: generate-test-keys
	go test -count=1 $(PKGS) -timeout=10m -coverprofile=coverage.txt -covermode=atomic ./...

license:
	@scripts/check_license.sh

lint:
	@scripts/check_lint.sh

checks: build license lint

generate-test-keys: clean
	@mkdir -p -p test/fixtures/keys/tls
	@docker run -i --rm \
		-v $(abspath .):/opt/go/src/github.com/trustbloc/hub-store \
		--entrypoint "/opt/go/src/github.com/trustbloc/hub-store/scripts/generate_test_keys.sh" \
		frapsoft/openssl

clean:
	rm -Rf ./test

all: checks unit-test
