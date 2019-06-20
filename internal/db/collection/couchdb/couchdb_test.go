/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/


package couchdb

import (
	"testing"

	"github.com/spf13/viper"

	"github.com/trustbloc/hub-store/tools/couchdb"

	"github.com/trustbloc/hub-store/internal/db/collection/commontests"
)

func TestStore(t *testing.T) {
	const dbname = "test"
	url, cleanup := couchdb.StartCouchDB(dbname)
	defer cleanup()
	err := couchdb.WaitForCouchDbStartup(url, dbname)
	if err != nil {
		panic(err)
	}
	cfg := viper.New()
	cfg.Set("collections.couchdbURL", url)
	cfg.Set("collections.couchdbName", dbname)
	commontests.RunTestsFor(t, Store(cfg))
}
