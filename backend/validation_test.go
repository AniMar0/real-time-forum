package backend

import "testing"

func TestIsValidNickname(t *testing.T) {
	tests := []struct {
		name     string
		nickname string
		valid    bool
	}{
		{name: "valid", nickname: "forum_user", valid: true},
		{name: "too short", nickname: "ab", valid: false},
		{name: "too long", nickname: "abcdefghijklmnopqrstu", valid: false},
		{name: "invalid character", nickname: "forum-user", valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isValidNickname(test.nickname); got != test.valid {
				t.Fatalf("isValidNickname(%q) = %v, want %v", test.nickname, got, test.valid)
			}
		})
	}
}

func TestIsValidPassword(t *testing.T) {
	if !isValidPassword("Secure123") {
		t.Fatal("expected a password with upper, lower, and number to be valid")
	}
	if isValidPassword("password") {
		t.Fatal("expected a password without required character classes to be invalid")
	}
}

func TestIsValidAge(t *testing.T) {
	if isValidAge(12) || isValidAge(121) {
		t.Fatal("expected ages outside the accepted range to be invalid")
	}
	if !isValidAge(13) || !isValidAge(120) {
		t.Fatal("expected boundary ages to be valid")
	}
}
