// Copyright 2025 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package rest

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

const (
	// type values for "typ" claim in token
	appType    = "app"
	deviceType = "device"
)

// holder for device and tenant id from token
type DeviceInfo struct {
	DeviceID string
	TenantID string
}

// holder to get extended claims that we expect
// device tokens and app tokens have differing claims
// but we take a union approach to keep it simple
type TokenClaims struct {
	TenantId string `json:"tid"`
	Type     string `json:"typ"`
	jwt.RegisteredClaims
}

// get bearer token and do common validation
func validateToken(r *http.Request) (*TokenClaims, error) {
	var claims TokenClaims

	accessToken, err := getBearerToken(r)
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(accessToken, &claims, getSigningKey)
	if err != nil {
		return nil, err
	} else if !token.Valid {
		return nil, ErrInvalidToken
	}

	if token.Header["alg"] == nil {
		return nil, ErrInvalidTokenHeaderSigningAlg
	}

	if !strings.HasPrefix(claims.Issuer, authConfig.Issuer) {
		return nil, ErrInvalidIssuerClaim
	}
	return &claims, nil
}

// do common validation and return claims for external facing apis
func getDeviceInfoFromToken(r *http.Request) (*DeviceInfo, error) {
	claims, err := validateToken(r)
	if err != nil {
		return nil, err
	}
	if claims.Type != deviceType {
		return nil, ErrInvalidTypeClaim
	}
	return &DeviceInfo{
		DeviceID: claims.Subject,
		TenantID: claims.TenantId,
	}, nil
}

func getBearerToken(r *http.Request) (string, error) {
	bearerToken := "Bearer "
	headerAuthorization := "Authorization"

	tokenString := r.Header.Get(headerAuthorization)
	if tokenString == "" {
		fsLogger.Error(ErrNoAuthorizationHeader.Error())
		return "", ErrNoAuthorizationHeader
	}
	if !strings.HasPrefix(tokenString, bearerToken) {
		fsLogger.Error(ErrNoBearerTokenSpecified.Error())
		return "", ErrNoBearerTokenSpecified
	}
	tokenString = strings.TrimPrefix(tokenString, bearerToken)
	return tokenString, nil
}
