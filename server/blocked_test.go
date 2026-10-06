package server

import (
	"errors"
	"testing"
)

func TestIsDeliveryForbidden(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "blocked by user", err: errors.New("Forbidden: bot was blocked by the user"), want: true},
		{name: "user deactivated", err: errors.New("Forbidden: user is deactivated"), want: true},
		{name: "generic forbidden", err: errors.New("Forbidden: bot can't initiate conversation with a user"), want: true},
		{name: "unrelated", err: errors.New("Bad Request: message is too long"), want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isDeliveryForbidden(tc.err); got != tc.want {
				t.Errorf("isDeliveryForbidden(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestRun_EnsuresUser(t *testing.T) {
	called := false
	db = &dbMock{
		ensureUserMock: func() error {
			called = true
			return nil
		},
	}

	cmd := newLocalCommand(42, "/help", &Options{})
	_ = cmd.run()

	if !called {
		t.Error("Expected EnsureUser to be called on inbound command")
	}
}
