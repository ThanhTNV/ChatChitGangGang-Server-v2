package auth

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey int

const (
	keyUserID ctxKey = iota
	keySubject
	keyEmail
	keyPreferredUsername
)

// WithPrincipal returns a child context carrying the authenticated principal.
func WithPrincipal(ctx context.Context, userID uuid.UUID, subject, email, preferredUsername string) context.Context {
	ctx = context.WithValue(ctx, keyUserID, userID)
	ctx = context.WithValue(ctx, keySubject, subject)
	ctx = context.WithValue(ctx, keyEmail, email)
	ctx = context.WithValue(ctx, keyPreferredUsername, preferredUsername)
	return ctx
}

// UserID returns the internal user UUID when present.
func UserID(ctx context.Context) (uuid.UUID, bool) {
	v, ok := ctx.Value(keyUserID).(uuid.UUID)
	return v, ok
}

// Subject returns the Keycloak subject (JWT "sub") when present.
func Subject(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(keySubject).(string)
	return v, ok
}

// Email returns the JWT email claim when present.
func Email(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(keyEmail).(string)
	return v, ok
}

// PreferredUsername returns the JWT preferred_username claim when present.
func PreferredUsername(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(keyPreferredUsername).(string)
	return v, ok
}
