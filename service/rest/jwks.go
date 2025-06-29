// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

const (
	// timeout for jwks http calls
	timeoutJwksGet = time.Second * time.Duration(5)
)

var (
	signingKeys = make(map[string]*rsa.PublicKey)
)

// jsonWebKey represents a JSON Web Key inside a JWKS.
type jsonWebKey struct {
	Curve    string `json:"crv"`
	Exponent string `json:"e"`
	K        string `json:"k"`
	ID       string `json:"kid"`
	Modulus  string `json:"n"`
	Type     string `json:"kty"`
	Use      string `json:"use"`
	X        string `json:"x"`
	Y        string `json:"y"`
}

// rawJWKS represents a JWKS in JSON format.
type rawJWKS struct {
	Keys []*jsonWebKey `json:"keys"`
}

// look up key by kid and validate token
// populates keys from jwks url into internal map as needed.
func getSigningKey(token *jwt.Token) (interface{}, error) {
	kid, ok := token.Header["kid"].(string)
	if !ok {
		fsLogger.Error("Invalid kid in the token header!")
		return nil, ErrInvalidTokenHeaderKid
	}

	// Check if a signing key corresponding to the kid was found in the
	// signing key table.
	pubKey, ok := signingKeys[kid]
	if !ok {
		// Key with this kid was not found - fetch the JWKS keys from the
		// configured jwks url to check if this is a new signing key.
		err := getJWKSSigningKey()
		if err != nil {
			fsLogger.Error("Failed to get JWKS signing keys!",
				zap.String("Token signed by:", kid),
				zap.Error(err),
			)
			return nil, err
		}
		pubKey, ok = signingKeys[kid]
		if !ok {
			return nil, fmt.Errorf("no public key to validate kid: %s", kid)
		}
	}

	return pubKey, nil
}

// Retrieve the token signing keys from JWKS endpoint.
func getJWKSSigningKey() error {
	url := authConfig.JwksUrl
	keys, err := getJwksFromServer(url)
	if err != nil {
		return err
	}
	if err = parseJWKS(keys); err != nil {
		fsLogger.Error("Error parsing keys.",
			zap.String("url:", url),
			zap.Error(err))
	}
	return err
}

// jwks requests
// adds a default timeout for http calls
func getJwksFromServer(url string) (keys []byte, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutJwksGet)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		fsLogger.Error("Error creating request for keys",
			zap.String("url", url),
			zap.Error(err))
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fsLogger.Error("Error fetching keys",
			zap.String("url", url),
			zap.Error(err))
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		fsLogger.Error("Get public keys failed",
			zap.String("url", url),
			zap.Int("status", resp.StatusCode),
			zap.Error(err))
		return nil, err
	}
	return io.ReadAll(resp.Body)
}

func parseJWKS(jwksBytes json.RawMessage) (err error) {
	var rawKS rawJWKS
	err = json.Unmarshal(jwksBytes, &rawKS)
	if err != nil {
		fsLogger.Error("Error unmarshalling jwks",
			zap.Error(err))
		return err
	}
	for _, key := range rawKS.Keys {
		switch keyType := key.Type; keyType {
		case ktyRSA:
			str, err := key.RSA()
			if err != nil {
				fsLogger.Error("Error parsing rsa key",
					zap.String("type:", key.Type),
					zap.String("kid:", key.ID),
					zap.Error(err))
				continue
			}
			signingKeys[key.ID] = str
		default:
			continue
		}
	}
	return nil
}
