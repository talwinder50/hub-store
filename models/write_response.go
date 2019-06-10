/*
   Copyright SecureKey Technologies Inc.
   This file contains software code that is the intellectual property of SecureKey.
   SecureKey reserves all rights in the code and you may not use it without written permission from SecureKey.
*/

package models

// WriteResponse write response
type WriteResponse struct {
	atContextField *string

	developerMessageField string

	// revisions
	// Required: true
	Revisions []string `json:"revisions"`
}
