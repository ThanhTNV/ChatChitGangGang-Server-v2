package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

func seedJWKSStore(t *testing.T, set jwk.Set) *JWKSStore {
	t.Helper()
	st := NewJWKSStore("http://unused.local/jwks", nil)
	st.mu.Lock()
	st.set = set
	st.fetched = time.Now()
	st.mu.Unlock()
	return st
}

func TestParseAndValidate_happyPath(t *testing.T) {
	t.Parallel()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privJWK, err := jwk.FromRaw(priv)
	if err != nil {
		t.Fatal(err)
	}
	if err := privJWK.Set(jwk.KeyIDKey, "test-kid"); err != nil {
		t.Fatal(err)
	}
	pubJWK, err := jwk.PublicKeyOf(privJWK)
	if err != nil {
		t.Fatal(err)
	}
	if err := pubJWK.Set(jwk.KeyIDKey, "test-kid"); err != nil {
		t.Fatal(err)
	}
	if err := pubJWK.Set(jwk.AlgorithmKey, jwa.RS256); err != nil {
		t.Fatal(err)
	}
	set := jwk.NewSet()
	if err := set.AddKey(pubJWK); err != nil {
		t.Fatal(err)
	}

	store := seedJWKSStore(t, set)
	v := NewValidator(store, "http://issuer.test/realms/demo", "api-client", time.Minute)

	tok, err := jwt.NewBuilder().
		Issuer("http://issuer.test/realms/demo").
		Subject("kc-sub-1").
		Audience([]string{"api-client"}).
		Expiration(time.Now().Add(time.Hour)).
		Claim("email", "a@b.c").
		Claim("preferred_username", "alice").
		Build()
	if err != nil {
		t.Fatal(err)
	}
	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256, privJWK))
	if err != nil {
		t.Fatal(err)
	}

	claims, err := v.ParseAndValidate(context.Background(), string(signed))
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != "kc-sub-1" || claims.Email != "a@b.c" || claims.PreferredUsername != "alice" {
		t.Fatalf("claims: %+v", claims)
	}
}

func TestParseAndValidate_azpAudience(t *testing.T) {
	t.Parallel()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privJWK, err := jwk.FromRaw(priv)
	if err != nil {
		t.Fatal(err)
	}
	if err := privJWK.Set(jwk.KeyIDKey, "kid"); err != nil {
		t.Fatal(err)
	}
	pubJWK, err := jwk.PublicKeyOf(privJWK)
	if err != nil {
		t.Fatal(err)
	}
	if err := pubJWK.Set(jwk.KeyIDKey, "kid"); err != nil {
		t.Fatal(err)
	}
	if err := pubJWK.Set(jwk.AlgorithmKey, jwa.RS256); err != nil {
		t.Fatal(err)
	}
	set := jwk.NewSet()
	if err := set.AddKey(pubJWK); err != nil {
		t.Fatal(err)
	}
	store := seedJWKSStore(t, set)
	v := NewValidator(store, "http://issuer", "flutter-public", time.Minute)

	tok, err := jwt.NewBuilder().
		Issuer("http://issuer").
		Subject("u1").
		Claim("azp", "flutter-public").
		Expiration(time.Now().Add(time.Hour)).
		Build()
	if err != nil {
		t.Fatal(err)
	}
	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256, privJWK))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v.ParseAndValidate(context.Background(), string(signed)); err != nil {
		t.Fatal(err)
	}
}
