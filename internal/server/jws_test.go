/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package server

import (
	"crypto/ecdsa"
	"io/ioutil"
	"log"
	"path/filepath"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/square/go-jose"
	"github.com/square/go-jose/jwt"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/hub-store/internal/crypto"
)

func TestCreateNewAccessToken(t *testing.T) {
	ecdsaKey := getPrivateKeyFromFile("../../test/bddtests/fixtures/keys/tls/ec-key.pem")
	_, err := createNewAccessToken("p6OLLpeRafCWbOAEYpuGVTKNkcq8l", "subject", ecdsaKey)
	require.NoError(t, err)
}

func TestCreateNewAccessTokenWithNilKey(t *testing.T) {
	var pk *ecdsa.PrivateKey
	_, err := createNewAccessToken("", "", pk)
	require.Error(t, err)
	require.Equal(t, "Failed to create new signer for new access token JWS: invalid private key", err.Error())
}

func TestValidateAccessToken(t *testing.T) {
	ecdsaKey := getPrivateKeyFromFile("../../test/bddtests/fixtures/keys/tls/ec-key.pem")
	tok, err := createNewAccessToken("p6OLLpeRafCWbOAEYpuGVTKNkcq8l", "subject", ecdsaKey)
	require.NoError(t, err)
	authJWT, parseErr := jwt.ParseSigned(tok)
	require.NoError(t, parseErr)
	err = validateAccessToken(authJWT, ecdsaKey, "subject")
	require.NoError(t, err)
}

func TestValidateAccessTokenWithWrongIssuer(t *testing.T) {
	ecdsaKey := getPrivateKeyFromFile("../../test/bddtests/fixtures/keys/tls/ec-key.pem")
	key := jose.SigningKey{Algorithm: jose.ES256, Key: ecdsaKey}
	var signerOpts = jose.SignerOptions{NonceSource: staticNonceSource("nonce")} // using passed in nonce
	signer, err := jose.NewSigner(key, signerOpts.WithType("JWT"))
	require.NoError(t, err)
	builder := jwt.Signed(signer)

	issuedTime := time.Now().UTC()
	expiryTime := issuedTime.Add(5 * time.Minute)
	claims := jwt.Claims{
		Issuer:    "Issuer",
		Subject:   "subject",
		ID:        "id",
		Audience:  jwt.Audience{HubIssuerID},
		NotBefore: jwt.NewNumericDate(issuedTime),
		IssuedAt:  jwt.NewNumericDate(issuedTime),
		Expiry:    jwt.NewNumericDate(expiryTime),
	}
	tok, err := builder.Claims(claims).CompactSerialize()
	require.NoError(t, err)
	authJWT, err := jwt.ParseSigned(tok)
	require.NoError(t, err)
	err = validateAccessToken(authJWT, ecdsaKey, "subject")
	require.Error(t, err)
	require.Equal(t, "Access Token validation failed: square/go-jose/jwt: validation failed, invalid issuer claim (iss)", err.Error())
}

func TestValidateAccessTokenWithWrongKey(t *testing.T) {
	ecdsaKey := getPrivateKeyFromFile("../../test/bddtests/fixtures/keys/server/ec-key.pem")
	tok, err := createNewAccessToken("p6OLLpeRafCWbOAEYpuGVTKNkcq8l", "subject", ecdsaKey)
	require.NoError(t, err)
	authJWT, parseErr := jwt.ParseSigned(tok)
	require.NoError(t, parseErr)

	ecdsaKey = getPrivateKeyFromFile("../../test/bddtests/fixtures/keys/client/ec-key.pem")
	err = validateAccessToken(authJWT, ecdsaKey, "subject")
	require.Error(t, err)
	require.Equal(t, "square/go-jose: error in cryptographic primitive", err.Error())
}

func getPrivateKeyFromFile(filePath string) *ecdsa.PrivateKey {
	keyBytes, err := ioutil.ReadFile(filepath.Clean(filePath))
	if err != nil {
		log.Panicf("Error while reading file = %s", filePath)
	}
	pvKey, err := crypto.ParsePrivateKey(keyBytes)
	if err != nil {
		log.Panicf("Error while parsing the contents of the file = %s", filePath)
	}
	ecdsaKey := pvKey.(*ecdsa.PrivateKey)
	return ecdsaKey
}

func TestRandomString(t *testing.T) {
	arr := make([]string, 1000)
	for i := 0; i < len(arr); i++ {
		arr[i] = randomString()
	}
	for i := 0; i < (len(arr) - 1); i++ {
		for j := i + 1; j < len(arr); j++ {
			require.NotEqual(t, arr[i], arr[j])
		}
	}
}

func TestValidateJWSHeader(t *testing.T) {
	kid := "did:example:123456789abcdefghi#keys-1"
	nonce := "p6OLLpeRafCWbOAEYpuGVTKNkcq8l"

	serverKey := getPrivateKeyFromFile("../../test/bddtests/fixtures/keys/server/ec-key.pem")
	clientKey := getPrivateKeyFromFile("../../test/bddtests/fixtures/keys/client/ec-key.pem")

	didAccessToken, err := createNewAccessToken(nonce, kid, serverKey)
	require.NoError(t, err)

	jwsMsg := generateValidJWS(kid, nonce, didAccessToken, clientKey)
	require.NoError(t, err)

	payload, err := validateJWSHeader(jwsMsg, &clientKey.PublicKey)
	require.NoError(t, err)
	require.NotNil(t, payload)
}

func TestValidateJWSHeaderWithInvalidKey(t *testing.T) {
	kid := "did:example:123456789abcdefghi#keys-1"
	nonce := "p6OLLpeRafCWbOAEYpuGVTKNkcq8l"

	serverKey := getPrivateKeyFromFile("../../test/bddtests/fixtures/keys/server/ec-key.pem")
	clientKey := getPrivateKeyFromFile("../../test//bddtests/fixtures/keys/client/ec-key.pem")

	didAccessToken, err := createNewAccessToken(nonce, kid, serverKey)
	require.NoError(t, err)

	jwsMsg := generateValidJWS(kid, nonce, didAccessToken, clientKey)
	require.NoError(t, err)

	payload, err := validateJWSHeader(jwsMsg, &serverKey.PublicKey)
	require.Error(t, err)
	require.Nil(t, payload)
	require.Equal(t, "Crypto [Warning]: could not verify JWS: square/go-jose: error in cryptographic primitive", err.Error())
}

func TestValidateJWSHeaderWithEmptyNonce(t *testing.T) {
	kid := "did:example:123456789abcdefghi#keys-1"
	nonce := ""

	serverKey := getPrivateKeyFromFile("../../test/bddtests/fixtures/keys/server/ec-key.pem")
	clientKey := getPrivateKeyFromFile("../../test/bddtests/fixtures/keys/client/ec-key.pem")

	didAccessToken, err := createNewAccessToken(nonce, kid, serverKey)
	require.NoError(t, err)

	jwsMsg := generateValidJWS(kid, nonce, didAccessToken, clientKey)
	require.NoError(t, err)

	payload, err := validateJWSHeader(jwsMsg, &clientKey.PublicKey)
	require.Error(t, err)
	require.Nil(t, payload)
	require.Equal(t, "Crypto [Warning]: Invalid token - missing nonce - %!s(<nil>)", err.Error())
}

func TestValidateJWS(t *testing.T) {
	kid := "did:example:123456789abcdefghi#keys-1"
	nonce := "p6OLLpeRafCWbOAEYpuGVTKNkcq8l"

	serverKey := getPrivateKeyFromFile("../../test/bddtests/fixtures/keys/server/ec-key.pem")
	clientKey := getPrivateKeyFromFile("../../test/bddtests/fixtures/keys/client/ec-key.pem")

	didAccessToken, err := createNewAccessToken(nonce, kid, serverKey)
	require.NoError(t, err)

	jwsMsg := generateValidJWS(kid, nonce, didAccessToken, clientKey)
	require.NoError(t, err)

	payload, err := validateJWSHeader(jwsMsg, &clientKey.PublicKey)
	require.NoError(t, err)
	require.NotNil(t, payload)

	token, isNew, err := validateJWS(jwsMsg, serverKey)
	require.NoError(t, err)
	require.False(t, isNew)
	require.True(t, token != "")
}

func TestValidateJWSWithoutAccessToken(t *testing.T) {
	kid := "did:example:123456789abcdefghi#keys-1"
	nonce := "p6OLLpeRafCWbOAEYpuGVTKNkcq8l"

	serverKey := getPrivateKeyFromFile("../../test/bddtests/fixtures/keys/server/ec-key.pem")
	clientKey := getPrivateKeyFromFile("../../test/bddtests/fixtures/keys/client/ec-key.pem")

	jwsMsg := generateValidJWS(kid, nonce, "", clientKey)

	payload, err := validateJWSHeader(jwsMsg, &clientKey.PublicKey)
	require.NoError(t, err)
	require.NotNil(t, payload)

	token, isNew, err := validateJWS(jwsMsg, serverKey)
	require.NoError(t, err)
	require.True(t, isNew)
	require.True(t, token != "")
}

func TestValidateJWSWithWrongAccessToken(t *testing.T) {
	kid := "did:example:123456789abcdefghi#keys-1"
	nonce := "p6OLLpeRafCWbOAEYpuGVTKNkcq8l"

	serverKey := getPrivateKeyFromFile("../../test/bddtests/fixtures/keys/server/ec-key.pem")
	clientKey := getPrivateKeyFromFile("../../test/bddtests/fixtures/keys/client/ec-key.pem")

	didAccessToken, err := createNewAccessToken(nonce, kid, serverKey)
	require.NoError(t, err)

	jwsMsg := generateValidJWS(kid, nonce, didAccessToken+".wrong", clientKey)
	require.NoError(t, err)

	payload, err := validateJWSHeader(jwsMsg, &clientKey.PublicKey)
	require.NoError(t, err)
	require.NotNil(t, payload)

	token, isNew, err := validateJWS(jwsMsg, serverKey)
	require.Error(t, err)
	require.False(t, isNew)
	require.Equal(t, "", token)
	require.Equal(t, "Crypto [Warning]: could not parse Access Token: square/go-jose: compact JWS format must have three parts", err.Error())
}

func TestValidateJWSWithWrongTime(t *testing.T) {
	kid := "did:example:123456789abcdefghi#keys-1"
	nonce := "p6OLLpeRafCWbOAEYpuGVTKNkcq8l"

	serverKey := getPrivateKeyFromFile("../../test/bddtests/fixtures/keys/server/ec-key.pem")
	clientKey := getPrivateKeyFromFile("../../test/bddtests/fixtures/keys/client/ec-key.pem")

	didAccessToken, err := createNewAccessTokenWithExpiredTime(nonce, kid, serverKey)
	require.NoError(t, err)

	jwsMsg := generateValidJWS(kid, nonce+"a", didAccessToken, clientKey)

	payload, err := validateJWSHeader(jwsMsg, &clientKey.PublicKey)
	require.NoError(t, err)
	require.NotNil(t, payload)

	token, isNew, err := validateJWS(jwsMsg, serverKey)
	require.Error(t, err)
	require.False(t, isNew)
	require.Equal(t, "", token)
	require.Equal(t, "Access Token validation failed: square/go-jose/jwt: validation failed, token is expired (exp)", err.Error())
}

func generateValidJWS(kid, nonce, accessToken string, clientKey *ecdsa.PrivateKey) *jose.JSONWebSignature {
	var err error
	headerAttrs := getJWSHeaderAttributes(kid, nonce, accessToken)
	jwsMsg := getJWS(headerAttrs, "This is a test payload.", "ES256", clientKey)
	jwsCompact, err := jwsMsg.CompactSerialize()
	if err != nil {
		log.Panicf("Error while compact serializing the jws: err=%s", err.Error())
	}

	jws, err := jose.ParseSigned(jwsCompact)
	if err != nil {
		log.Panicf("Error while parsing the signed jws")
	}

	return jws
}

func getJWSHeaderAttributes(kid, nonce, accessToken string) map[string]interface{} {
	headerAttrs := make(map[string]interface{})
	if kid != "" {
		headerAttrs[KidKey] = kid
	}
	if nonce != "" {
		headerAttrs[DidAccessNonceKey] = nonce
	}
	if accessToken != "" {
		headerAttrs[DidAccessTokenKey] = accessToken
	}
	return headerAttrs
}

// getJWS creates a JSON Web signature object from the payload
func getJWS(headerAttrs map[string]interface{}, payload string, algStr string, clientKey *ecdsa.PrivateKey) *jose.JSONWebSignature {
	alg := jose.SignatureAlgorithm(algStr)
	signerOpts := &jose.SignerOptions{}
	for k, v := range headerAttrs {
		signerOpts.WithHeader(jose.HeaderKey(k), v)
	}

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: alg, Key: clientKey}, signerOpts)
	if err != nil {
		log.Printf("Cannot get the Signer for the alg: %s", algStr)
	}
	// now sign a payload (JWT) to create a JWS and its serialized version
	jws, err := signer.Sign([]byte(payload))
	if err != nil {
		log.Panicf("Error while signing the payload")
	}
	return jws
}

func createNewAccessTokenWithExpiredTime(nonce, subject string, pvKey *ecdsa.PrivateKey) (string, error) {
	// for now creating keys with ECDSA using P-256 and SHA-256
	key := jose.SigningKey{Algorithm: jose.ES256, Key: pvKey}
	var signerOpts = jose.SignerOptions{NonceSource: staticNonceSource(nonce)} // using passed in nonce
	signerOpts.WithType("JWT")

	signer, err := jose.NewSigner(key, &signerOpts)
	if err != nil {
		return "", errors.Wrapf(err, "Failed to create new signer for new access token JWS")
	}

	builder := jwt.Signed(signer)

	issuedTime := time.Now().AddDate(0, 0, -1)
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
