// Package openai implements provider.Adapter for the OpenAI Chat
// Completions API (and compatible servers).
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/silverstreaminnovations/atlas/internal/provider"
)

const (
	defaultBaseURL = "https://api.openai.com"
	defaultModel   = "gpt-4o"
)

func init() {
	provider.Register("openai", func(s provider.Settings) provider.Adapter {
		return New(s)
	})
}

// Adapter talks to the OpenAI Chat Completions API.
type Adapter struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// New builds an OpenAI adapter, applying defaults for unset settings.
func New(s provider.Settings) *Adapter {
	a := &Adapter{apiKey: s.APIKey, baseURL: defaultBaseURL, model: defaultModel, client: http.DefaultClient}
	if s.BaseURL != "" {
		a.baseURL = s.BaseURL
	}
	if s.Model != "" {
		a.model = s.Model
	}
	return a
}

// Name implements provider.Adapter.
func (a *Adapter) Name() string { return "openai" }

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model     string        `json:"model"`
	MaxTokens int           `json:"max_tokens,omitempty"`
	Messages  []chatMessage `json:"messages"`
}

type choice struct {
	Message chatMessage `json:"message"`
}

type chatResponse struct {
	Model   string   `json:"model"`
	Choices []choice `json:"choices"`
}

// Complete implements provider.Adapter via POST /v1/chat/completions.
func (a *Adapter) Complete(ctx context.Context, req provider.Request) (provider.Response, error) {
	msgs := []chatMessage{}
	if req.System != "" {
		msgs = append(msgs, chatMessage{Role: "system", Content: req.System})
	}
	msgs = append(msgs, chatMessage{Role: "user", Content: req.Prompt})

	body, err := json.Marshal(chatRequest{Model: a.model, MaxTokens: req.MaxTokens, Messages: msgs})
	if err != nil {
		return provider.Response{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return provider.Response{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return provider.Response{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return provider.Response{}, fmt.Errorf("openai: unexpected status %s", resp.Status)
	}
	out := chatResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return provider.Response{}, err
	}
	if len(out.Choices) == 0 {
		return provider.Response{}, fmt.Errorf("openai: empty choices in response")
	}
	return provider.Response{Text: out.Choices[0].Message.Content, Model: out.Model}, nil
}

// Status implements provider.Adapter via GET /v1/models.
func (a *Adapter) Status(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+"/v1/models", nil)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("openai: unexpected status %s", resp.Status)
	}
	return nil
}
