package chat

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMessageCursorRoundTrip(t *testing.T) {
	t.Parallel()
	ts := time.Date(2026, 4, 10, 15, 30, 0, 123456789, time.UTC)
	id := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	s, err := encodeMessageCursor(ts, id)
	if err != nil {
		t.Fatal(err)
	}
	gotT, gotID, err := decodeMessageCursor(s)
	if err != nil {
		t.Fatal(err)
	}
	if !gotT.Equal(ts.UTC()) {
		t.Fatalf("time: got %v want %v", gotT, ts.UTC())
	}
	if gotID != id {
		t.Fatalf("id: got %v want %v", gotID, id)
	}
}
