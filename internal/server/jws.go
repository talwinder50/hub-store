/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package server

import (
	"crypto/ecdsa"
	"fmt"
	"math/rand"
	"time"

	"github.com/pkg/errors"
	"github.com/square/go-jose"
	"github.com/square/go-jose/jwt"
)

const (
	// DidAccessTokenKey is the JWS Header key for the DID Access Token
	DidAccessTokenKey = "did-access-token"

	// DidAccessNonceKey is the JWS Header key for the client's nonce
	DidAccessNonceKey = "did-requester-nonce"

	// KidKey is the did fragment which is pointing to the public key in the DID doc
	KidKey = "kid"

	// HubIssuerID represents the ID of the hub (TODO: in the future this should be made configurable for each instance)
	HubIssuerID = "did:hub:id"

	charset  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	idLength = 20
)

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

type staticNonceSource string

func (sns staticNonceSource) Nonce() (string, error) {
	return string(sns), nil
}

// createNewAccessToken will generate a new access token with the given nonce and using publicKey
func createNewAccessToken(nonce, subject string, pvKey *ecdsa.PrivateKey) (string, error) {
	// for now creating keys with ECDSA using P-256 and SHA-256
	key := jose.SigningKey{Algorithm: jose.ES256, Key: pvKey}
	var signerOpts = jose.SignerOptions{NonceSource: staticNonceSource(nonce)} // using passed in nonce
	signerOpts.WithType("JWT")

	signer, err := jose.NewSigner(key, &signerOpts)
	if err != nil {
		return "", errors.Wrapf(err, "Failed to create new signer for new access token JWS")
	}

	builder := jwt.Signed(signer)

	issuedTime := time.Now().UTC()
	expiryTime := issuedTime.Add(5 * time.Minute)
	claims := jwt.Claims{
		Issuer:    HubIssuerID,
		Subject:   subject,
		ID:        randomString(),
		Audience:  jwt.Audience{HubIssuerID},
		NotBefore: jwt.NewNumericDate(issuedTime),
		IssuedAt:  jwt.NewNumericDate(issuedTime),
		Expiry:    jwt.NewNumericDate(expiryTime),
	}

	return builder.Claims(claims).CompactSerialize()
}

func validateAccessToken(accessToken *jwt.JSONWebToken, pvKey *ecdsa.PrivateKey, subject string) error {
	resultClaims := jwt.Claims{}
	err := accessToken.Claims(&pvKey.PublicKey, &resultClaims)
	if err != nil {
		return err
	}

	err = resultClaims.Validate(jwt.Expected{
		Issuer:   HubIssuerID,
		Audience: jwt.Audience{HubIssuerID},
		Subject:  subject,
		ID:       resultClaims.ID,
		Time:     time.Now().UTC(),
	})
	if err != nil {
		return errors.Wrapf(err, "Access Token validation failed")
	}
	return nil
}

// randomString creates a random string of a constant length
func randomString() string {
	b := make([]byte, idLength)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// validateJWSHeader will validate the JWS header.
func validateJWSHeader(jws *jose.JSONWebSignature, publicKey *ecdsa.PublicKey) ([]byte, error) {
	_, _, verifiedPayload, err := jws.VerifyMulti(publicKey)
	if err != nil {
		return nil, errors.Wrapf(err, "Crypto [Warning]: could not verify JWS")
	}

	// validate nonce
	nonce, ok := jws.Signatures[0].Header.ExtraHeaders[jose.HeaderKey(DidAccessNonceKey)]
	if !ok || nonce == "" {
		return nil, errors.New(fmt.Sprintf("Crypto [Warning]: Invalid token - missing nonce - %s", nonce))
	}

	return verifiedPayload, nil
}

// validateJWS will validate the did-access-token in the JWS message. It will create a new token if not found.
// this call assumes validateJWSHeader() was called and returned a successful AuthenticationResult
func validateJWS(jws *jose.JSONWebSignature, key *ecdsa.PrivateKey) (string, bool, error) {
	kid := jws.Signatures[0].Header.KeyID

	accessTokenJwe := jws.Signatures[0].Header.ExtraHeaders[jose.HeaderKey(DidAccessTokenKey)]
	accessTknJweStr := fmt.Sprintf("%v", accessTokenJwe) // convert interface{} to string

	// for existing access token, return it back as is along with IsNewToken=false
	if accessTokenJwe != nil && accessTknJweStr != "" {
		authJWT, err := jwt.ParseSigned(accessTknJweStr)
		if err != nil {
			return "", false, errors.Wrapf(err, "Crypto [Warning]: could not parse Access Token")
		}

		err = validateAccessToken(authJWT, key, kid)
		if err != nil {
			return "", false, err
		}
		return accessTknJweStr, false, nil
	}

	// else create new token, return it along with IsNewToken=true
	nonce := jws.Signatures[0].Header.ExtraHeaders[jose.HeaderKey(DidAccessNonceKey)]

	accessTknJweStr, err := createNewAccessToken(fmt.Sprintf("%v", nonce), kid, key)
	if err != nil {
		return "", false, errors.Wrapf(err, "Crypto [Warning]: could not create Access Token")
	}

	return accessTknJweStr, true, nil
}
