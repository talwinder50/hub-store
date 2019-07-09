#!/bin/bash
#
# Copyright SecureKey Technologies Inc.
# This file contains software code that is the intellectual property of SecureKey.
# SecureKey reserves all rights in the code and you may not use it without written permission from SecureKey.
#

# This script runs tests.
# It accepts one arg - the go build tags to test.

set -e

# Packages to exclude
PKGS=`go list github.com/trustbloc/hub-store/... 2> /dev/null | \
                                                  grep -v /test | \
                                                  grep -v /vendor/`
echo "Running unit tests ..."
go test -count=1 -cover $PKGS -timeout=10m
