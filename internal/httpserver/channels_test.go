package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/chatchitganggang/internal-comm-backend/internal/auth"
	"github.com/chatchitganggang/internal-comm-backend/internal/channel"
)

type stubChannelStore struct {
	list      []channel.Channel
	listErr   error
	createErr error
	createID  uuid.UUID
}

func (s *stubChannelStore) ListForUser(ctx context.Context, userID uuid.UUID) ([]channel.Channel, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.list, nil
}

func (s *stubChannelStore) CreateGroup(ctx context.Context, userID uuid.UUID, name string) (channel.Channel, error) {
	if s.createErr != nil {
		return channel.Channel{}, s.createErr
	}
	id := s.createID
	if id == uuid.Nil {
		id = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	}
	cb := userID
	return channel.Channel{
		ID:        id,
		Name:      name,
		Type:      "group",
		CreatedBy: &cb,
		CreatedAt: time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC),
	}, nil
}

func TestListChannels_OK(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	chID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	ts := time.Date(2026, 4, 6, 12, 0, 0, 0, time.UTC)
	stub := &stubChannelStore{
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
	stub := &stubChannelStore{listErr: context.Canceled}
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

func TestCreateChannel_OK(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	stub := &stubChannelStore{}
	h := createChannelHandler(stub, log)
	payload := `{"name":"general","type":"group"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/channels", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(auth.WithPrincipal(req.Context(), uid, "sub", "", ""))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	var ch channel.Channel
	if err := json.NewDecoder(rec.Body).Decode(&ch); err != nil {
		t.Fatal(err)
	}
	if ch.Name != "general" || ch.Type != "group" || ch.CreatedBy == nil || *ch.CreatedBy != uid {
		t.Fatalf("unexpected channel: %+v", ch)
	}
}

func TestCreateChannel_defaultTypeGroup(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	uid := uuid.New()
	stub := &stubChannelStore{}
	h := createChannelHandler(stub, log)
	req := httptest.NewRequest(http.MethodPost, "/v1/channels", strings.NewReader(`{"name":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(auth.WithPrincipal(req.Context(), uid, "sub", "", ""))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d", rec.Code)
	}
}

func TestCreateChannel_rejectDirect(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	stub := &stubChannelStore{}
	h := createChannelHandler(stub, log)
	req := httptest.NewRequest(http.MethodPost, "/v1/channels", strings.NewReader(`{"name":"dm","type":"direct"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(auth.WithPrincipal(req.Context(), uuid.New(), "sub", "", ""))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestCreateChannel_emptyName(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := createChannelHandler(&stubChannelStore{}, log)
	req := httptest.NewRequest(http.MethodPost, "/v1/channels", strings.NewReader(`{"name":"  ","type":"group"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(auth.WithPrincipal(req.Context(), uuid.New(), "sub", "", ""))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestCreateChannel_nameTooLong(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	longName := strings.Repeat("あ", maxChannelNameRunes+1)
	body, _ := json.Marshal(map[string]string{"name": longName, "type": "group"})
	h := createChannelHandler(&stubChannelStore{}, log)
	req := httptest.NewRequest(http.MethodPost, "/v1/channels", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(auth.WithPrincipal(req.Context(), uuid.New(), "sub", "", ""))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestCreateChannel_internalError(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	stub := &stubChannelStore{createErr: context.Canceled}
	h := createChannelHandler(stub, log)
	req := httptest.NewRequest(http.MethodPost, "/v1/channels", strings.NewReader(`{"name":"x","type":"group"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(auth.WithPrincipal(req.Context(), uuid.New(), "sub", "", ""))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", rec.Code)
	}
}

func TestCreateChannel_nilStore(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := createChannelHandler(nil, log)
	req := httptest.NewRequest(http.MethodPost, "/v1/channels", strings.NewReader(`{"name":"x"}`))
	req = req.WithContext(auth.WithPrincipal(req.Context(), uuid.New(), "sub", "", ""))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", rec.Code)
	}
}
