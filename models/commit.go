/*
   Copyright SecureKey Technologies Inc.
   This file contains software code that is the intellectual property of SecureKey.
   SecureKey reserves all rights in the code and you may not use it without written permission from SecureKey.
*/

package models

// Commit commit
type Commit struct {

	// header
	Header *Header `json:"header,omitempty"`

	// payload
	Payload Payload `json:"payload,omitempty"`

	// protected
	Protected *Protected `json:"protected,omitempty"`

	// signature
	Signature string `json:"signature,omitempty"`
}
