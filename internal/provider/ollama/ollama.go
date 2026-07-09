// Package ollama implements provider.Adapter for a local Ollama server.
package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Haykhay/atlas/internal/provider"
)

const (
	defaultBaseURL = "http://localhost:11434"
	defaultModel   = "llama3.2"
)

func init() {
	provider.Register("ollama", func(s provider.Settings) provider.Adapter {
		return New(s)
	})
}

// Adapter talks to an Ollama server over its native HTTP API.
type Adapter struct {
	baseURL string
	model   string
	client  *http.Client
}

// New builds an Ollama adapter, applying defaults for unset settings.
func New(s provider.Settings) *Adapter {
	a := &Adapter{baseURL: defaultBaseURL, model: defaultModel, client: http.DefaultClient}
	if s.BaseURL != "" {
		a.baseURL = s.BaseURL
	}
	if s.Model != "" {
		a.model = s.Model
	}
	return a
}

// Name implements provider.Adapter.
func (a *Adapter) Name() string { return "ollama" }

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Model   string      `json:"model"`
	Message chatMessage `json:"message"`
}

// Complete implements provider.Adapter via POST /api/chat.
func (a *Adapter) Complete(ctx context.Context, req provider.Request) (provider.Response, error) {
	msgs := []chatMessage{}
	if req.System != "" {
		msgs = append(msgs, chatMessage{Role: "system", Content: req.System})
	}
	msgs = append(msgs, chatMessage{Role: "user", Content: req.Prompt})

	body, err := json.Marshal(chatRequest{Model: a.model, Messages: msgs, Stream: false})
	if err != nil {
		return provider.Response{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return provider.Response{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return provider.Response{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return provider.Response{}, fmt.Errorf("ollama: unexpected status %s", resp.Status)
	}
	out := chatResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return provider.Response{}, err
	}
	return provider.Response{Text: out.Message.Content, Model: out.Model}, nil
}

// Status implements provider.Adapter via GET /api/tags.
func (a *Adapter) Status(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+"/api/tags", nil)
	if err != nil {
		return err
	}
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama: unexpected status %s", resp.Status)
	}
	return nil
}
