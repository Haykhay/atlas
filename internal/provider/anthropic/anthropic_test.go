package anthropic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/silverstreaminnovations/atlas/internal/provider"
)

func TestCompleteSendsMessagesRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "sk-test" {
			t.Errorf("missing api key header")
		}
		if r.Header.Get("anthropic-version") == "" {
			t.Errorf("missing anthropic-version header")
		}
		var req messagesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if req.MaxTokens == 0 {
			t.Error("max_tokens must default to non-zero")
		}
		if req.System != "sys" || req.Messages[0].Content != "hello" {
			t.Errorf("request not forwarded: %+v", req)
		}
		_ = json.NewEncoder(w).Encode(messagesResponse{
			Model:   "claude-sonnet-5",
			Content: []contentBlock{{Type: "text", Text: "world"}},
		})
	}))
	defer srv.Close()

	a := New(provider.Settings{APIKey: "sk-test", BaseURL: srv.URL})
	resp, err := a.Complete(context.Background(), provider.Request{System: "sys", Prompt: "hello"})
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if resp.Text != "world" || resp.Model != "claude-sonnet-5" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestStatusChecksModels(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" || r.Header.Get("x-api-key") != "sk-test" {
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
