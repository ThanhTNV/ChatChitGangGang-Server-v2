package httpserver

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/chatchitganggang/internal-comm-backend/internal/auth"
	"github.com/chatchitganggang/internal-comm-backend/internal/channel"
)

type stubChannelLister struct {
	list []channel.Channel
	err  error
}

func (s *stubChannelLister) ListForUser(ctx context.Context, userID uuid.UUID) ([]channel.Channel, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.list, nil
}

func TestListChannels_OK(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	chID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	ts := time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC)
	stub := &stubChannelLister{
		list: []channel.Channel{
			{ID: chID, Name: "general", Type: "group", CreatedAt: ts},
		},
	}
	h := listChannelsHandler(stub, log)
	req := httptest.NewRequest(http.MethodGet, "/v1/channels", nil)
	req = req.WithContext(auth.WithPrincipal(req.Context(), uid, "sub", "", ""))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Channels []channel.Channel `json:"channels"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Channels) != 1 || body.Channels[0].ID != chID || body.Channels[0].Name != "general" {
		t.Fatalf("unexpected payload: %+v", body)
	}
}

func TestListChannels_internalError(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	stub := &stubChannelLister{err: context.Canceled}
	h := listChannelsHandler(stub, log)
	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	req := httptest.NewRequest(http.MethodGet, "/v1/channels", nil)
	req = req.WithContext(auth.WithPrincipal(req.Context(), uid, "sub", "", ""))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", rec.Code)
	}
}

func TestListChannels_nilRepo(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := listChannelsHandler(nil, log)
	req := httptest.NewRequest(http.MethodGet, "/v1/channels", nil)
	req = req.WithContext(auth.WithPrincipal(req.Context(), uuid.New(), "sub", "", ""))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", rec.Code)
	}
}
