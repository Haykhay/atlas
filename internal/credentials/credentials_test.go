package credentials

import "testing"

func TestSetGetDeleteRoundTrip(t *testing.T) {
	MockForTesting()

	if err := Set("anthropic", "sk-test"); err != nil {
		t.Fatalf("set: %v", err)
	}
	got, err := Get("anthropic")
	if err != nil || got != "sk-test" {
		t.Fatalf("get: %q, %v", got, err)
	}
	if err := Delete("anthropic"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := Get("anthropic"); !IsNotFound(err) {
		t.Fatalf("expected not-found, got %v", err)
	}
}
