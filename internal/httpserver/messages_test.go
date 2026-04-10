package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/chatchitganggang/internal-comm-backend/internal/auth"
	"github.com/chatchitganggang/internal-comm-backend/internal/chat"
)

type stubMessageStore struct {
	res chat.ListMessagesResult
	err error
}

func (s *stubMessageStore) ListForMember(ctx context.Context, userID, channelID uuid.UUID, opts chat.ListMessagesOpts) (chat.ListMessagesResult, error) {
	return s.res, s.err
}

func TestListChannelMessages_OK(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	chID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	msgID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	ts := time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC)
	stub := &stubMessageStore{
		res: chat.ListMessagesResult{
			Messages: []chat.Message{
				{
					ID:        msgID,
					ChannelID: chID,
					SenderID:  uid,
					Body:      json.RawMessage(`{"text":"hi"}`),
					CreatedAt: ts,
				},
			},
			NextCursor: "next-page-token",
		},
	}
	r := chi.NewRouter()
	r.Get("/v1/channels/{channelID}/messages", listChannelMessagesHandler(stub, log))
	req := httptest.NewRequest(http.MethodGet, "/v1/channels/"+chID.String()+"/messages", nil)
	req = req.WithContext(auth.WithPrincipal(req.Context(), uid, "sub", "", ""))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Messages   []chat.Message `json:"messages"`
		NextCursor string         `json:"next_cursor"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Messages) != 1 || body.Messages[0].ID != msgID {
		t.Fatalf("payload %+v", body)
	}
	if body.NextCursor != "next-page-token" {
		t.Fatalf("next_cursor %q", body.NextCursor)
	}
}

func TestListChannelMessages_notMember(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	stub := &stubMessageStore{err: chat.ErrNotMember}
	r := chi.NewRouter()
	r.Get("/v1/channels/{channelID}/messages", listChannelMessagesHandler(stub, log))
	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	chID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	req := httptest.NewRequest(http.MethodGet, "/v1/channels/"+chID.String()+"/messages", nil)
	req = req.WithContext(auth.WithPrincipal(req.Context(), uid, "sub", "", ""))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", rec.Code)
	}
}

func TestListChannelMessages_invalidChannelID(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	stub := &stubMessageStore{}
	r := chi.NewRouter()
	r.Get("/v1/channels/{channelID}/messages", listChannelMessagesHandler(stub, log))
	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	req := httptest.NewRequest(http.MethodGet, "/v1/channels/not-a-uuid/messages", nil)
	req = req.WithContext(auth.WithPrincipal(req.Context(), uid, "sub", "", ""))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestListChannelMessages_invalidLimit(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	stub := &stubMessageStore{}
	r := chi.NewRouter()
	r.Get("/v1/channels/{channelID}/messages", listChannelMessagesHandler(stub, log))
	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	chID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	req := httptest.NewRequest(http.MethodGet, "/v1/channels/"+chID.String()+"/messages?limit=abc", nil)
	req = req.WithContext(auth.WithPrincipal(req.Context(), uid, "sub", "", ""))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestListChannelMessages_invalidCursor(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	stub := &stubMessageStore{err: fmt.Errorf("%w: %v", chat.ErrInvalidCursor, errors.New("decode"))}
	r := chi.NewRouter()
	r.Get("/v1/channels/{channelID}/messages", listChannelMessagesHandler(stub, log))
	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	chID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	req := httptest.NewRequest(http.MethodGet, "/v1/channels/"+chID.String()+"/messages", nil)
	req = req.WithContext(auth.WithPrincipal(req.Context(), uid, "sub", "", ""))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestListChannelMessages_nilStore(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	r := chi.NewRouter()
	r.Get("/v1/channels/{channelID}/messages", listChannelMessagesHandler(nil, log))
	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	chID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	req := httptest.NewRequest(http.MethodGet, "/v1/channels/"+chID.String()+"/messages", nil)
	req = req.WithContext(auth.WithPrincipal(req.Context(), uid, "sub", "", ""))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", rec.Code)
	}
}
