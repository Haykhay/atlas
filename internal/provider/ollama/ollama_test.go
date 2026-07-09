package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Haykhay/atlas/internal/provider"
)

func TestCompleteSendsChatRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Errorf("path: %s", r.URL.Path)
		}
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if req.Stream {
			t.Error("stream must be false")
		}
		if req.Messages[len(req.Messages)-1].Content != "hello" {
			t.Errorf("prompt not forwarded: %+v", req.Messages)
		}
		_ = json.NewEncoder(w).Encode(chatResponse{
			Model:   "llama3.2",
			Message: chatMessage{Role: "assistant", Content: "world"},
		})
	}))
	defer srv.Close()

	a := New(provider.Settings{BaseURL: srv.URL})
	resp, err := a.Complete(context.Background(), provider.Request{System: "sys", Prompt: "hello"})
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if resp.Text != "world" || resp.Model != "llama3.2" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestStatusChecksTags(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	a := New(provider.Settings{BaseURL: srv.URL})
	if err := a.Status(context.Background()); err != nil {
		t.Fatalf("status: %v", err)
	}
}
