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


#Release Parameters
BASE_VERSION = 0.3.2
IS_RELEASE = false

GO_CMD ?= go
export GO111MODULE=on


#couchdb image parameters
#this couchdb image contains startup scripts that autorun and create all necessary db artifacts
export HUB_STORE_COUCHDB_IMAGE ?= trustbloc/hub-store-couchdb

export PKGS=`go list github.com/trustbloc/hub-store/... `
ifneq ($(IS_RELEASE),true)
EXTRA_VERSION ?= snapshot-$(shell git rev-parse --short=7 HEAD)
PROJECT_VERSION=$(BASE_VERSION)-$(EXTRA_VERSION)
ARTIFACTORY_PORT=8443
else
PROJECT_VERSION=$(BASE_VERSION)
ARTIFACTORY_PORT=8444
endif

# This Makefile assumes a working Golang and Docker setup
export ARCH=$(shell go env GOARCH)
DOCKER_CMD         ?= docker
CONTAINER_IDS = $(shell docker ps -a -q)
DEV_IMAGES = $(shell docker images dev-* -q)

#couchdb image parameters
#this couchdb image contains startup scripts that autorun and create all necessary db artifacts
export HUB_STORE_COUCHDB_IMAGE ?= repo.onetap.ca:$(ARTIFACTORY_PORT)/next/trustbloc/hub-store-couchdb
export HUB_STORE_COUCHDB_IMAGE_TAG ?= $(PROJECT_VERSION)

docker:
	GOOS=linux go build -o ./build/hub-store cmd/hub-store/main.go
	@docker build --no-cache --tag $(HUB_STORE_COUCHDB_IMAGE):$(ARCH)-$(HUB_STORE_COUCHDB_IMAGE_TAG) \
	--tag $(HUB_STORE_COUCHDB_IMAGE):$(ARCH)-latest ./scripts/couchdb

ifdef JENKINS_URL
ifndef JENKINS_VERIFY
    docker push $(HUB_STORE_COUCHDB_IMAGE):$(ARCH)-$(HUB_STORE_COUCHDB_IMAGE_TAG)
endif
endif


build:
	go build -o bin/hubstore cmd/hub-store/main.go

//TODO: Pull the couchdb image directly , dont build the image to run the test
docker:
	@docker build --no-cache --tag $(HUB_STORE_COUCHDB_IMAGE) \
	./scripts/couchdb

unit-test:depend docker
	@scripts/unit.sh
//TODO : Separate the couchdb test (which are dependant on external dependencies ) as integration test
unit-test: generate-test-keys docker
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
