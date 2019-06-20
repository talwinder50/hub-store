/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package couchdb

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/trustbloc/hub-store/internal"

	"github.com/spf13/viper"
	"github.com/trustbloc/hub-store/internal/db/collection"

	"github.com/pkg/errors"

	"github.com/trustbloc/hub-store/internal/db"

	_ "github.com/go-kivik/couchdb" // The CouchDB driver
	"github.com/go-kivik/kivik"     // Development version of Kivik
	"github.com/trustbloc/hub-store/internal/server/models"
)

// Store returns a CouchDB implementation of db.Store.
func Store(config *viper.Viper) collection.Store {
	const urlKey = "collections.couchdbURL"
	const dbnameKey = "collections.couchdbName"
	const errMsg = "failed to initialize couchdb collections.Store: key %s not found"

	url := config.GetString(urlKey)
	if len(url) == 0 {
		panic(fmt.Errorf(errMsg, urlKey))
	}
	dbname := config.GetString(dbnameKey)
	if len(dbname) == 0 {
		panic(fmt.Errorf(errMsg, dbnameKey))
	}
	client, err := kivik.New("couch", url)
	if err != nil {
		panic(err)
	}
	db, err := client.DB(context.TODO(), dbname)
	if err != nil {
		panic(err)
	}
	return &couchDB{db: db}
}

// We store commits inside this envelope in order to pin each one to an object ID.
type envelope struct {
	ObjectID  string         `json:"objectID"`
	Interface string         `json:"interface"`
	Context   string         `json:"context"`
	Tpe       string         `json:"type"`
	MetaName  string         `json:"name"`
	Commit    *models.Commit `json:"commit"`
}

// couchDB implements db.Store.
type couchDB struct {
	// CouchDB client
	db *kivik.DB
}

func (c *couchDB) Write(commit *models.Commit) error {
	oid, e := internal.ObjectID(commit)
	if e != nil {
		return e
	}
	iface, ctx, tpe, name, e := parseCommit(commit)
	if e != nil {
		return e
	}
	_, _, err := c.db.CreateDoc(
		context.TODO(),
		&envelope{
			ObjectID:  oid,
			Commit:    commit,
			Interface: iface,
			Context:   ctx,
			Tpe:       tpe,
			MetaName:  name},
	)
	return err
}
func parseCommit(commit *models.Commit) (string, string, string, string, error) {

	protected := &models.Protected{}
	protectedBytes, err := decoding(commit.Protected)
	if err != nil {
		return "", "", "", "", err
	}
	err = json.Unmarshal(protectedBytes, protected)
	if err != nil {
		return "", "", "", "", err
	}
	iface := protected.Interface
	ctx := protected.Context
	tpe := protected.Type
	name := protected.Meta.Name

	return iface, ctx, tpe, name, nil

}
func decoding(input string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(input)
}
func (c *couchDB) ObjectQuery(iface, ctx, tpe string, f *collection.Filter, p *db.Paging) (metadata []*models.Commit, next string, err error) {
	var params map[string]interface{}
	params, err = objectQueryParams(iface, ctx, tpe, f, p)
	if err != nil {
		return nil, "", &collection.ErrUnsupportedFilter{Msg: errors.Wrapf(
			err,
			"failed to construct ObjectQuery parameters for interface=%s, context=%s, type=%s, filter=%+v, paging=%+v",
			iface, ctx, tpe, f, p,
		).Error()}
	}
	var rows *kivik.Rows
	rows, err = c.db.Find(context.TODO(), params)
	if err != nil {
		return nil, "", errors.Wrapf(
			err,
			"failed to execute ObjectQuery on CouchDB for interface=%s context=%s type=%s params=%+v",
			iface, ctx, tpe, params,
		)
	}
	metadata, next, err = fetchCommits(rows, params)
	if err != nil {
		return nil, "", errors.Wrapf(
			err,
			"failed to fetch ObjectQuery metadata from CouchDB for interface=%s context=%s type=%s params=%+v",
			iface, ctx, tpe, params,
		)
	}
	return metadata, next, nil
}

func (c *couchDB) CommitQuery(oid string, f *collection.Filter, p *db.Paging) (commits []*models.Commit, next string, err error) {
	var params map[string]interface{}
	params, err = commitQueryParams(oid, f, p)
	if err != nil {
		return nil, "", errors.Wrapf(err,
			"failed to construct CommitQuery parameters for oid=%s, filter=%+v, paging=%+v",
			oid, f, p,
		)
	}
	var rows *kivik.Rows
	rows, err = c.db.Find(context.TODO(), params)
	if err != nil {
		return nil, "", errors.Wrapf(
			err,
			"failed to execute CommitQuery on CouchDB for oid=%s params=%v",
			oid, params)
	}
	commits, next, err = fetchCommits(rows, params)
	if err != nil {
		return nil, "", errors.Wrapf(
			err,
			"failed to fetch CommitQuery commits from CouchDB for oid=%s params=%+v",
			oid, params,
		)
	}
	return commits, next, nil
}

// Load commits from CouchDB.
// Pagination is achieved by querying CouchDB for pageSize + 1 rows. We return up to 'pageSize' results back to the user;
// if CouchDB returns more rows than this then we also return a skip_token to the user.
func fetchCommits(rows *kivik.Rows, params map[string]interface{}) (commits []*models.Commit, token string, err error) {
	commits = make([]*models.Commit, 0)
	limit, limited := params["limit"]
	if !limited {
		limit = math.MaxInt64
	}
	for i := 1; rows.Next(); i++ {
		if i == limit.(int) {
			skip, skipping := params["skip"]
			if !skipping {
				skip = 0
			}
			token = strconv.Itoa(skip.(int) + limit.(int) - 1)
		} else {
			env := envelope{}
			if err = rows.ScanDoc(&env); err != nil {
				return nil, "", errors.Wrapf(
					err, "failed to unmarshal envelope doc with key [%s]", rows.Key(),
				)
			}
			commitBytes, err := json.Marshal(env.Commit)
			if err != nil {
				return nil, "", errors.Wrapf(
					err, "failed to marshal envelope commit to bytes",
				)
			}
			envCommits := &models.Commit{}
			err = json.Unmarshal(commitBytes, envCommits)
			if err != nil {
				return nil, "", errors.Wrapf(
					err, "failed to unmarshal envelope commit to models.commit",
				)
			}

			commits = append(commits, envCommits)
		}
	}
	return commits, token, nil
}
