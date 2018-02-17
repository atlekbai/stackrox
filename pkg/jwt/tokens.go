// Package jwt parses and validates JSON Web Tokens (JWTs).
package jwt

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"bitbucket.org/stack-rox/apollo/pkg/logging"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

// A Validator checks JSON Web Tokens (JWTs) to ensure they are intended for
// this service and are cryptographically trusted.
type Validator interface {
	Validate(hdrs map[string][]string) (token string, claims *jwt.Claims, err error)
}

type rs256Validator struct {
	keyGetter KeyGetter
	expected  jwt.Expected
}

// NewRS256Validator validates tokens generated using RS256 (256-bit RSA).
func NewRS256Validator(keys KeyGetter, issuer, audience string) Validator {
	return rs256Validator{
		keyGetter: keys,
		expected:  jwt.Expected{Issuer: issuer, Audience: []string{audience}},
	}
}

var (
	// ErrNoToken means no token was provided.
	ErrNoToken = errors.New("no token provided")
	// ErrNoKeyID means that no key ID was provided, so validation could not be completed.
	ErrNoKeyID = errors.New("no key ID provided")
	// ErrKeyNotFound means that the referenced key was not in the list of known keys.
	ErrKeyNotFound = errors.New("referenced key not found")
	// ErrNoJWTHeaders means that there were no headers in the JWT (and therefore no signatures to verify).
	ErrNoJWTHeaders = errors.New("no headers found in the JWT")
	// ErrInvalidAlgorithm means that the token uses an algorithm not valid for this validator.
	ErrInvalidAlgorithm = errors.New("invalid algorithm used")
	// ErrUnverifiableToken means that, despite all efforts, the token could not be validated.
	ErrUnverifiableToken = errors.New("token could not be validated")

	logger = logging.LoggerForModule()
)

// tokenFromHeader looks for the token in the Authorization header.
func tokenFromHeader(hdrs map[string][]string) (string, *jwt.JSONWebToken, error) {
	raw := fromHeader(hdrs["authorization"]) // gRPC metadata keys are lowercased.
	if raw == nil {
		return "", nil, ErrNoToken
	}
	t, err := jwt.ParseSigned(string(raw))
	return string(raw), t, err
}

func fromHeader(hdrs []string) []byte {
	if len(hdrs) == 0 {
		return nil
	}
	hdr := hdrs[0] // Disregard repeated settings for the header.
	if len(hdr) > 7 && strings.EqualFold(hdr[0:7], "BEARER ") {
		return []byte(hdr[7:])
	}
	return nil
}

// Validate validates the token or returns an error.
func (v rs256Validator) Validate(hdrs map[string][]string) (string, *jwt.Claims, error) {
	raw, token, err := tokenFromHeader(hdrs)
	if err != nil {
		return "", nil, err
	}

	if len(token.Headers) < 1 {
		return "", nil, ErrNoJWTHeaders
	}

	for _, h := range token.Headers {
		claims, err := v.validateWithHeader(token, h)
		if err == nil {
			return raw, claims, nil
		}
		logger.Warnf("Validation error: %s", err)
	}
	return "", nil, ErrUnverifiableToken
}

func (v rs256Validator) validateWithHeader(token *jwt.JSONWebToken, header jose.Header) (*jwt.Claims, error) {
	if header.Algorithm != string(jose.RS256) {
		return nil, ErrInvalidAlgorithm
	}

	if header.KeyID == "" {
		return nil, ErrNoKeyID
	}
	key := v.keyGetter.Key(header.KeyID)
	if key == nil {
		return nil, ErrKeyNotFound
	}

	var claims jwt.Claims
	err := token.Claims(key, &claims)
	if err != nil {
		return nil, fmt.Errorf("claim processing: %s", err)
	}

	err = claims.Validate(v.expected.WithTime(time.Now()))
	if err != nil {
		return nil, err
	}
	return &claims, err
}
