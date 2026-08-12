package backend

import "testing"

func TestHubSupportsMultipleSessionsPerUser(t *testing.T) {
	hub := NewHub()
	first := &Client{ID: "first", UserID: 1, Username: "alice", Send: make(chan interface{}, 1)}
	second := &Client{ID: "second", UserID: 1, Username: "alice", Send: make(chan interface{}, 1)}

	hub.Register(first)
	hub.Register(second)

	if !hub.IsOnline(1) {
		t.Fatal("expected alice to be online")
	}
	if got := len(hub.ClientsForUser(1)); got != 2 {
		t.Fatalf("got %d active sessions, want 2", got)
	}

	hub.SendToUser(1, "event")
	if got := len(first.Send); got != 1 {
		t.Fatalf("first session received %d events, want 1", got)
	}
	if got := len(second.Send); got != 1 {
		t.Fatalf("second session received %d events, want 1", got)
	}

	hub.Unregister(first)
	if !hub.IsOnline(1) {
		t.Fatal("alice should remain online while the second session is connected")
	}
	hub.Unregister(second)
	if hub.IsOnline(1) {
		t.Fatal("alice should be offline after both sessions disconnect")
	}
}
