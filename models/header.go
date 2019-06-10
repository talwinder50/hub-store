/*
   Copyright SecureKey Technologies Inc.
   This file contains software code that is the intellectual property of SecureKey.
   SecureKey reserves all rights in the code and you may not use it without written permission from SecureKey.
*/

package models

// Header header
type Header struct {

	// iss
	Iss string `json:"iss,omitempty"`

	// rev
	Rev string `json:"rev,omitempty"`
}
