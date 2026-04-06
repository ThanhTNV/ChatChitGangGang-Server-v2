package auth

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
)

// JWKSStore fetches and caches a remote JWKS document.
type JWKSStore struct {
	url    string
	client *http.Client
	ttl    time.Duration

	mu      sync.RWMutex
	set     jwk.Set
	fetched time.Time
}

// NewJWKSStore builds a store for the given JWKS HTTPS/HTTP URL.
func NewJWKSStore(url string, client *http.Client) *JWKSStore {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &JWKSStore{
		url:    url,
		client: client,
		ttl:    10 * time.Minute,
	}
}

// Get returns the cached set or refreshes from the network.
func (s *JWKSStore) Get(ctx context.Context, forceRefresh bool) (jwk.Set, error) {
	if !forceRefresh {
		s.mu.RLock()
		if s.set != nil && time.Since(s.fetched) < s.ttl {
			set := s.set
			s.mu.RUnlock()
			return set, nil
		}
		s.mu.RUnlock()
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if !forceRefresh && s.set != nil && time.Since(s.fetched) < s.ttl {
		return s.set, nil
	}

	set, err := jwk.Fetch(ctx, s.url, jwk.WithHTTPClient(s.client))
	if err != nil {
		return nil, fmt.Errorf("fetch jwks: %w", err)
	}
	s.set = set
	s.fetched = time.Now()
	return set, nil
}

// Invalidate drops the cache so the next Get performs a fetch.
func (s *JWKSStore) Invalidate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.set = nil
}
