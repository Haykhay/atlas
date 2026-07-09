package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/silverstreaminnovations/atlas/internal/provider"
)

func TestCompleteSendsChatCompletion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer sk-test" {
			t.Errorf("missing bearer token")
		}
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if req.Messages[len(req.Messages)-1].Content != "hello" {
			t.Errorf("prompt not forwarded: %+v", req.Messages)
		}
		_ = json.NewEncoder(w).Encode(chatResponse{
			Model:   "gpt-4o",
			Choices: []choice{{Message: chatMessage{Role: "assistant", Content: "world"}}},
		})
	}))
	defer srv.Close()

	a := New(provider.Settings{APIKey: "sk-test", BaseURL: srv.URL})
	resp, err := a.Complete(context.Background(), provider.Request{System: "sys", Prompt: "hello"})
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if resp.Text != "world" || resp.Model != "gpt-4o" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestStatusChecksModels(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" || r.Header.Get("Authorization") != "Bearer sk-test" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	a := New(provider.Settings{APIKey: "sk-test", BaseURL: srv.URL})
	if err := a.Status(context.Background()); err != nil {
		t.Fatalf("status: %v", err)
	}
}
