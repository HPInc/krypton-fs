package rest

import (
	"errors"
)

var (
	ErrUrlDecode                    = errors.New("error decoding URL path")
	ErrPayloadRead                  = errors.New("payload read error")
	ErrAuthz                        = errors.New("authorization error")
	ErrPathVariableMissing          = errors.New("a required path variable was not specified")
	ErrInvalidToken                 = errors.New("invalid token provided")
	ErrInvalidTokenHeaderKid        = errors.New("invalid token signing kid specified")
	ErrInvalidTokenHeaderSigningAlg = errors.New("invalid token signing algorithm specified")
	ErrInvalidIssuerClaim           = errors.New("specified token contains an invalid issuer claim")
	ErrInvalidAudienceClaim         = errors.New("specified token contains an invalid audience claim")
	ErrInvalidSubjectClaim          = errors.New("specified token contains an invalid subject claim")
	ErrInvalidTypeClaim             = errors.New("specified token contains an invalid typ claim")
	ErrNoAuthorizationHeader        = errors.New("request does not have an authorization header")
	ErrNoBearerTokenSpecified       = errors.New("authorization header does not contain a bearer token")
)
