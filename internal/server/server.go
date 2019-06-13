/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package server

import "github.com/trustbloc/hub-store/internal/db/collection"

// Config is a holder for the identity hub's configuration.
type Config struct {
	// PageSize configures the page size to use for paged queries.
	PageSize int
	// Collection store.
	CollectionStore collection.Store
}
