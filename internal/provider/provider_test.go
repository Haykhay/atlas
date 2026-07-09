package provider

import (
	"context"
	"testing"
)

type fakeAdapter struct{ settings Settings }

func (f *fakeAdapter) Name() string { return "fake" }
func (f *fakeAdapter) Complete(_ context.Context, _ Request) (Response, error) {
	return Response{Text: "hi", Model: f.settings.Model}, nil
}
func (f *fakeAdapter) Status(_ context.Context) error { return nil }

func TestRegistryNewAndNames(t *testing.T) {
	Register("fake", func(s Settings) Adapter { return &fakeAdapter{settings: s} })

	found := false
	for _, n := range Names() {
		if n == "fake" {
			found = true
		}
	}
	if !found {
		t.Fatalf("Names() missing registered provider: %v", Names())
	}

	a, err := New("fake", Settings{Model: "m1"})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	resp, err := a.Complete(context.Background(), Request{Prompt: "x"})
	if err != nil || resp.Model != "m1" {
		t.Fatalf("complete: %+v, %v", resp, err)
	}

	if _, err := New("nope", Settings{}); err == nil {
		t.Fatal("expected error for unknown provider")
	}
}
