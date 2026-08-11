package account

import (
	"context"
	"testing"
)

func TestIdentityContextRoundTrip(t *testing.T) {
	expected := Identity{Nickname: "alice", SessionID: "session-1"}
	ctx := WithIdentity(context.Background(), expected)

	got, ok := IdentityFromContext(ctx)
	if !ok {
		t.Fatal("expected identity to be present")
	}
	if got != expected {
		t.Fatalf("got %#v, want %#v", got, expected)
	}
}
