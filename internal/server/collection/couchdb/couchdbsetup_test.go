/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package couchdb

import (
	"context"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-kivik/couchdb" // The CouchDB driver
	"github.com/go-kivik/kivik"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func Pull(image string, timeout time.Duration) error {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return errors.Wrap(err, "cannot create docker client")
	}
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	err = client.PullImage(
		docker.PullImageOptions{
			Context: ctx,
			Repository: strings.Replace(image, ":", "@", 1),
		},
		docker.AuthConfiguration{},
	)
	if err != nil {
		return errors.Wrapf(err, "cannot pull image %s", image)
	}
	return nil
}

// StartCouchDB starts a docker container and returns its address.
// Use the cleanup function to stop it.
func StartCouchDB(image string, timeout time.Duration) (address string, cleanup func()) {
	dockerClient, err := docker.NewClientFromEnv()
	if err != nil {
		panic(err)
	}
	cdb := &couchDB{
		Name:          uuid.New().String(),
		Image:         image,
		HostIP:        "127.0.0.1",
		ContainerPort: docker.Port("5984/tcp"),
		StartTimeout:  timeout,
		Client:        dockerClient,
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

func CreateDB(couchDbUrl, couchDbName string, timeout time.Duration) error {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	client, err := cdbClient(couchDbUrl, ctx)
	_, err = client.CreateDB(ctx, couchDbName)
	if err != nil {
		return errors.Wrapf(err, "cannot create couchdb %s", couchDbName)
	}
	return nil
}

func CreateIndices(couchDbUrl, couchDbName string, timeout time.Duration) error {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	client, err := cdbClient(couchDbUrl, ctx)
	if err != nil {
		return err
	}
	db, err := client.DB(ctx, couchDbName)
	if err != nil {
		return errors.Wrap(err, "cannot obtain handle to couchdb")
	}
	err = db.CreateIndex(
		ctx, "hub", "objectquery",
		map[string]interface{}{
			"partial_filter_selector": map[string]string{
				"commit.protected.operation": "create",
			},
			"fields": []string{
				"commit.protected.interface",
				"commit.protected.context",
				"commit.protected.type",
			},
		},
	)
	if err != nil {
		return errors.Wrap(err, "cannot create objectquery index")
	}
	err = db.CreateIndex(
		ctx, "hub", "commitquery",
		map[string]interface{}{
			"fields":[]string{"objectID"},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "cannot create commitquery index")
	}
	return nil
}

func cdbClient(url string, ctx context.Context) (*kivik.Client, error) {
	client, err := kivik.New("couch", url)
	if err != nil {
		return nil, errors.Wrap(err, "cannot initialize couchdb client")
	}
	return client, nil
}
