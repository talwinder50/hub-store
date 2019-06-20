/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package couchdb

import (
	"strconv"

	"github.com/trustbloc/hub-store/internal/db"
	"github.com/trustbloc/hub-store/internal/db/collection"
)

// Query parameters for the objectquery index.
// This index is expected to have pre-filtered all the "create" commits.
// We always match on 'interface' AND 'context' AND 'type'. If oids are provided,
// then we further filter the results to only those objects with the given IDs.
func objectQueryParams(iface, context, tpe string, filter *collection.Filter, paging *db.Paging) (params map[string]interface{}, err error) {
	selector := make(map[string]interface{})
	selector["interface"] = iface
	selector["context"] = context
	selector["type"] = tpe
	if len(filter.Oids) > 0 || len(filter.MetadataFilters) > 0 {
		and := make([]map[string]interface{}, 0)
		if len(filter.Oids) > 0 {
			or := make([]map[string]interface{}, 0)
			for _, oid := range filter.Oids {
				or = append(or, map[string]interface{}{"objectID": oid})
			}
			and = append(
				and,
				map[string]interface{}{"$or": or},
			)
		}
		if len(filter.MetadataFilters) > 0 {
			for _, f := range filter.MetadataFilters {
				var condition map[string]interface{}
				condition, err = asCondition(f)
				if err != nil {
					return nil, err
				}
				and = append(and, condition)
			}
		}
		selector["$and"] = and
	}
	params = map[string]interface{}{
		"selector":  selector,
		"use_index": []string{"hub", "objectquery"},
	}
	if err = skipTokenParam(paging, params); err != nil {
		return nil, err
	}
	pageSizeParam(paging, params)
	return params, nil
}

// Query parameters for the commitquery index.
// We match on all objects with the given oid. If revisions are provided,
// we filter the results further to only commits with matching revs.
func commitQueryParams(oid string, filter *collection.Filter, paging *db.Paging) (params map[string]interface{}, err error) {
	selector := map[string]interface{}{
		"objectID": oid,
	}
	if len(filter.Revs) > 0 {
		or := make([]map[string]interface{}, len(filter.Revs))
		for i, rev := range filter.Revs {
			or[i] = map[string]interface{}{"commit.header.rev": rev}
		}
		selector["$or"] = or
	}
	params = map[string]interface{}{
		"selector":  selector,
		"use_index": []string{"hub", "commitquery"},
	}
	if err = skipTokenParam(paging, params); err != nil {
		return nil, err
	}
	pageSizeParam(paging, params)
	return params, nil
}

func skipTokenParam(options *db.Paging, params map[string]interface{}) error {
	if len(options.SkipToken) > 0 {
		skip, err := strconv.Atoi(options.SkipToken)
		if err != nil {
			return err
		}
		params["skip"] = skip
	}
	return nil
}

// Pagination achieved by querying for pageSize + 1 results. Up to 'pageSize' results are returned to
// the user. If the query returns more, then we return a skip_token to the user.
func pageSizeParam(options *db.Paging, params map[string]interface{}) {
	if options.Size > 0 {
		params["limit"] = options.Size + 1
	}
}
