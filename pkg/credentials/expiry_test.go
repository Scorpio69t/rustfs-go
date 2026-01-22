package credentials

import (
	"testing"
	"time"
)

func TestExpirySetExpirationDefaultWindow(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	expiration := now.Add(10 * time.Second)

	exp := &Expiry{CurrentTime: func() time.Time { return now }}
	exp.SetExpiration(expiration, DefaultExpiryWindow)

	exp.CurrentTime = func() time.Time { return now.Add(7 * time.Second) }
	if exp.IsExpired() {
		t.Fatalf("expected credentials to be valid before the cutoff")
	}

	exp.CurrentTime = func() time.Time { return now.Add(9 * time.Second) }
	if !exp.IsExpired() {
		t.Fatalf("expected credentials to be expired after the cutoff")
	}
}

func TestExpirySetExpirationCustomWindow(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	expiration := now.Add(10 * time.Second)

	exp := &Expiry{CurrentTime: func() time.Time { return now }}
	exp.SetExpiration(expiration, 3*time.Second)

	exp.CurrentTime = func() time.Time { return now.Add(7 * time.Second) }
	if exp.IsExpired() {
		t.Fatalf("expected credentials to be valid at the cutoff")
	}

	exp.CurrentTime = func() time.Time { return now.Add(8 * time.Second) }
	if !exp.IsExpired() {
		t.Fatalf("expected credentials to be expired after the cutoff")
	}
}
