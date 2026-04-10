package chat

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type messageCursorPayload struct {
	T time.Time `json:"t"`
	I uuid.UUID `json:"i"`
}

func encodeMessageCursor(createdAt time.Time, id uuid.UUID) (string, error) {
	p := messageCursorPayload{T: createdAt.UTC(), I: id}
	raw, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func decodeMessageCursor(s string) (time.Time, uuid.UUID, error) {
	if s == "" {
		return time.Time{}, uuid.Nil, fmt.Errorf("empty cursor")
	}
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return time.Time{}, uuid.Nil, fmt.Errorf("cursor decode: %w", err)
	}
	var p messageCursorPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return time.Time{}, uuid.Nil, fmt.Errorf("cursor json: %w", err)
	}
	if p.I == uuid.Nil {
		return time.Time{}, uuid.Nil, fmt.Errorf("cursor missing id")
	}
	return p.T.UTC(), p.I, nil
}
