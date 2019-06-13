/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package collection

import (
	"github.com/trustbloc/hub-store/internal/db"
	"github.com/trustbloc/hub-store/internal/server/models"
)

// Store defines methods for interacting with the backing db for the identity hub.
type Store interface {
	// Write writes the commit to the store.
	Write(commit *models.Commit) error

	// CommitQuery returns all commits (or the subset specified with the 'Revs' or 'CommitOp' filters) for a
	// single object.
	//
	// Pagination is achieved with a combination of paging options 'PageSize', 'SkipToken' and the 'token' return
	// parameter.
	//
	// 'token' is an opaque token to be used to fetch a subsequent page, if there's any.
	CommitQuery(oid string, f *Filter, p *db.Paging) (commits []*models.Commit, token string, err error)
}

// Filter is a set of criteria that narrows down the results of query methods.
type Filter struct {
	// Constrain the search to these object IDs.
	Oids []string
	// Constrain the search to these revision IDs.
	Revs []string
	// ObjectQuery filters that match on a commit's metadata.
	MetadataFilters []*models.Filter
}
