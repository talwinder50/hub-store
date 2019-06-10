/*
   Copyright SecureKey Technologies Inc.
   This file contains software code that is the intellectual property of SecureKey.
   SecureKey reserves all rights in the code and you may not use it without written permission from SecureKey.
*/

package models

// BaseRequest base request
type BaseRequest struct {
	atContextField *string

	atTypeField string

	audField *string

	issField *string

	subField *string
}
