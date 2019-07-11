#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# Command used to start couchdb.
# This was taken from the parent image here: https://github.com/apache/couchdb-docker/blob/master/2.2.0/Dockerfile
START_COUCHDB=/opt/couchdb/bin/couchdb
COUCHDB_PID=0

function waitForCouchDB() {
	max_retries=5
	attempts=1
	while [ $attempts -le $max_retries  ]; do
		response=$(curl -o /dev/null --silent -w "%{http_code}" http://localhost:5984)
		if [ $response == "200" ]; then
			break
		fi
		if [ $attempts -ge $max_retries ]; then
			echo "maxed out wait time for CouchDB startup - shutting down"
			exit 1
		fi
		sleep 3
		let attempts=$attempts+1
	done
}

function main() {
	trap 'kill -SIGTERM $COUCHDB_PID; wait $COUCHDB_PID' SIGTERM
	# Start CouchDB in the background
	$START_COUCHDB &
	COUCHDB_PID="$!"
	waitForCouchDB
	./setup_couchdb.sh
	if [ $? -ne 0 ]; then
		echo "error during couchdb setup - shutting down"
		exit $?
	fi
	wait $COUCHDB_PID
}

main
