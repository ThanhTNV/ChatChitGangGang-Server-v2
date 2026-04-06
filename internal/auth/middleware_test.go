package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

type fakeSync struct {
	id uuid.UUID
}

func (f fakeSync) UpsertByKeycloakSub(ctx context.Context, keycloakSub, displayName string) (uuid.UUID, error) {
	if f.id != uuid.Nil {
		return f.id, nil
	}
	return uuid.MustParse("11111111-1111-1111-1111-111111111111"), nil
}

func TestBearerMiddleware_missingToken(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h := BearerMiddleware(&Validator{}, fakeSync{})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next called")
	}))
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
}
