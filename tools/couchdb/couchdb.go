/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package couchdb

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/go-kivik/couchdb" // The CouchDB driver
	"github.com/go-kivik/kivik"     // Development version of Kivik

	"net/http"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/thedevsaddam/gojsonq"
)

func couchdbImageName() string {
	return os.Getenv("HUB_STORE_COUCHDB_IMAGE") + ":" + os.Getenv("ARCH") + "-latest"
}

// StartCouchDB starts a CouchDB test instance and returns its address.
// Use the cleanup function to stop it.
func StartCouchDB(dbname string) (address string, cleanup func()) {
	dockerClient, err := docker.NewClientFromEnv()
	if err != nil {
		panic(err)
	}
	cdb := &couchDB{
		Name:          uuid.New().String(),
		Image:         couchdbImageName(),
		HostIP:        "127.0.0.1",
		ContainerPort: docker.Port("5984/tcp"),
		StartTimeout:  60 * time.Second,
		Client:        dockerClient,
		Env:           []string{"COUCHDB_DBNAME=" + dbname},
	}
	if err := cdb.Start(); err != nil {
		fmtErr := fmt.Errorf("failed to start couchDB: %s", err)
		panic(fmtErr)
	}
	return cdb.Address(), func() {
		err := cdb.Stop()
		if err != nil {
			panic(err)
		}
	}
}

// CreateView creates a view on the CouchDB database dbname hosted on the given url.
// 'ddoc' and 'viewName' may or may not be prefixed with "_design/" and "_view/" respectively.
// The reduce function is optional.
func CreateView(url, dbname, ddoc, viewName, mapf string, reducef ...string) {
	ddoc = "_design/" + strings.TrimPrefix(ddoc, "_design/")
	viewName = strings.TrimPrefix(viewName, "_view/")
	view := make(map[string]interface{})
	view["map"] = mapf
	if len(reducef) > 0 {
		view["reduce"] = reducef[0]
	}
	_, err := newCouchDbClient(url, dbname).Put(
		context.TODO(), ddoc,
		map[string]interface{}{
			"views": map[string]interface{}{
				viewName: view,
			},
		},
	)
	if err != nil {
		panic(err)
	}
}

// Create a new CouchDB client.
func newCouchDbClient(url, dbname string) *kivik.DB {
	client, err := kivik.New("couch", url)
	if err != nil {
		panic(err)
	}
	db, err := client.DB(context.TODO(), dbname)
	if err != nil {
		panic(err)
	}
	return db
}

// WaitForCouchDbStartup waits for the CouchDB docker container to start up
// Our test CouchDB docker container starts up quickly but takes a few seconds to finish configuring
// itself. We want to avoid running the tests before CouchDB is ready.
// This function tries to ping the CouchDB hosted at the given URL a few times, returning an error
// for any connectivity issue or if the maximum timeout was reached.
func WaitForCouchDbStartup(couchDbURL, couchDbName string) error {
	indexes := []string{"commitquery", "objectquery", "grants"}
	const timeout = 60 * time.Second
	limit := time.Now().Add(timeout)
	indexStatus := make([]bool, len(indexes))
	for time.Now().Before(limit) && !logicalAnd(indexStatus) {
		resp, err := http.Get(fmt.Sprintf("http://%s/%s/_index", couchDbURL, couchDbName))
		if err != nil {
			return errors.Wrapf(err, "error pinging CouchDB test url: %s", couchDbURL)
		}
		json := gojsonq.New().Reader(resp.Body)
		if err := resp.Body.Close(); err != nil {
			return errors.Wrapf(err, "failed to close http response from CouchDB")
		}
		for i, index := range indexes {
			indexStatus[i] = json.From("indexes").Where("name", "=", index).Get() != nil
			json.Reset()
		}
		time.Sleep(1 * time.Second)
	}
	if !logicalAnd(indexStatus) {
		return errors.New("CouchDB startup timed out")
	}
	return nil
}

func logicalAnd(status []bool) bool {
	result := true
	for _, b := range status {
		if !b {
			result = false
			break
		}
	}
	return result
}
