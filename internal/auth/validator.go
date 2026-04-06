package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwt"
)

// Claims holds OIDC claims used by the API after validation.
type Claims struct {
	Subject             string
	Email               string
	PreferredUsername   string
	AuthorizedParty     string // azp
}

// Validator verifies JWT access tokens against JWKS and configured issuer/audience rules.
type Validator struct {
	store    *JWKSStore
	issuer   string
	audience string
	skew     time.Duration
}

// NewValidator constructs a validator. audience is required; empty issuer should not be passed.
func NewValidator(store *JWKSStore, issuer, audience string, skew time.Duration) *Validator {
	if skew <= 0 {
		skew = 30 * time.Second
	}
	return &Validator{
		store:    store,
		issuer:   issuer,
		audience: audience,
		skew:     skew,
	}
}

// ParseAndValidate parses a compact JWT, verifies signature via JWKS, and checks issuer, exp, and audience/azp.
func (v *Validator) ParseAndValidate(ctx context.Context, raw string) (*Claims, error) {
	if v.store == nil {
		return nil, fmt.Errorf("auth: nil jwks store")
	}
	parseOnce := func(forceJWKS bool) (jwt.Token, error) {
		set, err := v.store.Get(ctx, forceJWKS)
		if err != nil {
			return nil, err
		}
		tok, err := jwt.ParseString(
			raw,
			jwt.WithKeySet(set),
			jwt.WithIssuer(v.issuer),
			jwt.WithValidate(true),
			jwt.WithAcceptableSkew(v.skew),
		)
		if err != nil {
			return nil, err
		}
		if !audienceMatches(tok, v.audience) {
			return nil, fmt.Errorf("jwt: aud/azp does not match expected audience")
		}
		return tok, nil
	}

	tok, err := parseOnce(false)
	if err != nil {
		if strings.Contains(err.Error(), "aud/azp does not match") {
			return nil, err
		}
		v.store.Invalidate()
		tok, err = parseOnce(true)
		if err != nil {
			return nil, err
		}
	}

	return claimsFromToken(tok), nil
}

func audienceMatches(tok jwt.Token, want string) bool {
	if want == "" {
		return true
	}
	if azp, ok := tok.Get("azp"); ok {
		if s, _ := azp.(string); s == want {
			return true
		}
	}
	switch aud := tok.Audience(); len(aud) {
	case 0:
		return false
	default:
		for _, a := range aud {
			if a == want {
				return true
			}
		}
		return false
	}
}

func claimsFromToken(tok jwt.Token) *Claims {
	c := &Claims{
		Subject: tok.Subject(),
	}
	if v, ok := tok.Get("email"); ok {
		c.Email, _ = v.(string)
	}
	if v, ok := tok.Get("preferred_username"); ok {
		c.PreferredUsername, _ = v.(string)
	}
	if v, ok := tok.Get("azp"); ok {
		c.AuthorizedParty, _ = v.(string)
	}
	return c
}
