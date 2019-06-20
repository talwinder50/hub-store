/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package couchdb

import (
	"fmt"

	"github.com/trustbloc/hub-store/internal/server/models"
)

// Return a CouchDB condition - to be used inside a selector - based on the filter.
func asCondition(filter *models.Filter) (condition map[string]interface{}, err error) {
	operators := map[string]func(*models.Filter) map[string]interface{}{
		"eq": eqCondition,
	}
	for id, op := range operators {
		if id == filter.Type {
			condition = op(filter)
			break
		}
	}
	if condition == nil {
		err = fmt.Errorf("filter.type not supported for filter %+v", filter)
	}
	return condition, err
}

// A CouchDB condition based on equality of the metadata field named by
// the filter to the value also specified by the filter.
func eqCondition(filter *models.Filter) map[string]interface{} {
	return map[string]interface{}{
		filter.Field: filter.Value,
	}
}
