#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# This script sets up an already running CouchDB instance for use by the hub-store.
# It requires:
#   - A running instance of CouchDB v2.2.0
#   - cURL
#
# Environment variables that affect this script:
# COUCHDB_URL:    	full url to the running CouchDB instance
# COUCHDB_USER:   	CouchDB username (optional)
# COUCHDB_PASSWORD: CouchDB password (optional)
# COUCHDB_DBNAME: 	desired name of the database to create

set -e

HTTP_OK="'200'"
HTTP_CREATED="'201'"

# Temporary file to store session cookie in case we need to authenticate.
COOKIE=cookie
# Options for cURL
CURL_OUTPUT=curl.out
CURL_COMMON_OPTIONS="--silent -o ${CURL_OUTPUT} -w '%{http_code}'"

# Cleans up temporary files
function cleanup() {
	rm $CURL_OUTPUT
	if [ -e $COOKIE ]; then
		rm $COOKIE
	fi
}

# Generic function to check the result of a cURL command against the expected value
function checkResult() {
	EXPECTED=$1
	ACTUAL=$2
	MSG=$3
	if [ $ACTUAL != $EXPECTED ]; then
		echo $MSG
		echo "CouchDB response code: $ACTUAL"
		echo -n "CouchDB response msg: " | cat - $CURL_OUTPUT
		cleanup
		exit 1
	fi
}

# Creates a session by authenticating with COUCHDB_USER and COUCHDB_PASSWORD
function createSession() {
	echo "Authenticating..."
	result=$(curl $CURL_COMMON_OPTIONS -c $COOKIE -H "Content-Type: application/x-www-form-urlencoded" -d "name=$COUCHDB_USER&password=$COUCHDB_PASSWORD" $COUCHDB_URL/_session)
	checkResult $HTTP_OK $result "could not authenticate!"
}

# Creates the database named COUCHDB_DBNAME
function createDatabase() {
	echo "Creating database named [$COUCHDB_DBNAME]..."
	result=$($CURL -X PUT $COUCHDB_URL/$COUCHDB_DBNAME)
	checkResult $HTTP_CREATED $result "could not create database!"
}

# Creates the indexes for the collections store.
function createCollectionsStoreIndexes() {
	echo "Creating design document [hub] with indexes [commitquery, objectquery] for the collections store..."
	objectquery=$(cat <<- EOF
	{
		 "index": {
			 "partial_filter_selector": {
				"commit.protected.operation": "create"
			 },
			 "fields": [
				"commit.protected.interface",
				"commit.protected.context",
				"commit.protected.type"
			 ]
		 },
		 "name": "objectquery",
		 "ddoc": "hub",
		 "type": "json"
	}
	EOF
	)
	commitquery=$(cat <<- EOF
	{
		"index": {
			"fields": ["objectID"]
		},
		"name": "commitquery",
		"ddoc": "hub",
		"type": "json"
	}
	EOF
	)
	result=$($CURL -X POST -H "Content-Type: application/json" -d "$objectquery" $COUCHDB_URL/$COUCHDB_DBNAME/_index)
	checkResult $HTTP_OK $result "could not create the objectquery index!"
	result=$($CURL -X POST -H "Content-Type: application/json" -d "$commitquery" $COUCHDB_URL/$COUCHDB_DBNAME/_index)
	checkResult $HTTP_OK $result "could not create the commitquery index!"
}

# Creates the indexes required by the identity-hub
function createIndexes() {
    createCollectionsStoreIndexes
}

# Main function
function main() {
	if [ -z "$COUCHDB_PASSWORD" ]; then
		CURL="curl $CURL_COMMON_OPTIONS"
	else
		createSession
		CURL="curl -b $COOKIE $CURL_COMMON_OPTIONS"
	fi
	
	createDatabase
	createIndexes
	cleanup
	exit 0
}

main
