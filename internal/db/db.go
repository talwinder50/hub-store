/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package db

// Paging is a generic set of options that control paging in store queries that support it.
type Paging struct {
	// Page size.
	Size int
	// Set the SkipToken to the one returned by the store in order to fetch the next page
	// of results.
	SkipToken string
}
