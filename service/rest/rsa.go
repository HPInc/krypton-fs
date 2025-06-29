// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

/*
This file is from https://github.com/MicahParks/keyfunc
License: Apache v2.0
  - https://github.com/MicahParks/keyfunc/blob/master/LICENSE
*/
package rest

import (
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

const (
	// ktyRSA is the key type (kty) in the JWT header for RSA.
	ktyRSA = "RSA"
)

var (
	ErrMissingAssets = errors.New("required assets are missing to create a public key")
)

// RSA parses a jsonWebKey and turns it into an RSA public key.
func (j *jsonWebKey) RSA() (publicKey *rsa.PublicKey, err error) {
	if j.Exponent == "" || j.Modulus == "" {
		return nil, fmt.Errorf("%w: %s", ErrMissingAssets, ktyRSA)
	}

	// Decode the exponent from Base64.
	//
	// According to RFC 7518, this is a Base64 URL unsigned integer.
	// https://tools.ietf.org/html/rfc7518#section-6.3
	exponent, err := base64urlTrailingPadding(j.Exponent)
	if err != nil {
		return nil, err
	}
	modulus, err := base64urlTrailingPadding(j.Modulus)
	if err != nil {
		return nil, err
	}

	// Turn the exponent into an integer.
	//
	// According to RFC 7517, these numbers are in big-endian format.
	// https://tools.ietf.org/html/rfc7517#appendix-A.1
	return &rsa.PublicKey{
		E: int(big.NewInt(0).SetBytes(exponent).Uint64()),
		N: big.NewInt(0).SetBytes(modulus),
	}, nil
}

// base64urlTrailingPadding removes trailing padding before decoding a string from base64url. Some non-RFC compliant
// JWKS contain padding at the end values for base64url encoded public keys.
//
// Trailing padding is required to be removed from base64url encoded keys.
// RFC 7517 defines base64url the same as RFC 7515 Section 2:
// https://datatracker.ietf.org/doc/html/rfc7517#section-1.1
// https://datatracker.ietf.org/doc/html/rfc7515#section-2
func base64urlTrailingPadding(s string) ([]byte, error) {
	s = strings.TrimRight(s, "=")
	return base64.RawURLEncoding.DecodeString(s)
}
