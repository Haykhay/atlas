// Package anthropic implements provider.Adapter for the Anthropic
// Messages API.
package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Haykhay/atlas/internal/provider"
)

const (
	defaultBaseURL   = "https://api.anthropic.com"
	defaultModel     = "claude-sonnet-5"
	defaultMaxTokens = 4096
	apiVersion       = "2023-06-01"
)

func init() {
	provider.Register("anthropic", func(s provider.Settings) provider.Adapter {
		return New(s)
	})
}

// Adapter talks to the Anthropic Messages API.
type Adapter struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// New builds an Anthropic adapter, applying defaults for unset settings.
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
func (a *Adapter) Name() string { return "anthropic" }

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type messagesRequest struct {
	Model     string        `json:"model"`
	MaxTokens int           `json:"max_tokens"`
	System    string        `json:"system,omitempty"`
	Messages  []chatMessage `json:"messages"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type messagesResponse struct {
	Model   string         `json:"model"`
	Content []contentBlock `json:"content"`
}

func (a *Adapter) headers(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", apiVersion)
}

// Complete implements provider.Adapter via POST /v1/messages.
func (a *Adapter) Complete(ctx context.Context, req provider.Request) (provider.Response, error) {
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = defaultMaxTokens
	}
	body, err := json.Marshal(messagesRequest{
		Model:     a.model,
		MaxTokens: maxTokens,
		System:    req.System,
		Messages:  []chatMessage{{Role: "user", Content: req.Prompt}},
	})
	if err != nil {
		return provider.Response{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return provider.Response{}, err
	}
	a.headers(httpReq)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return provider.Response{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return provider.Response{}, fmt.Errorf("anthropic: unexpected status %s", resp.Status)
	}
	out := messagesResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return provider.Response{}, err
	}
	texts := []string{}
	for _, block := range out.Content {
		if block.Type == "text" {
			texts = append(texts, block.Text)
		}
	}
	return provider.Response{Text: strings.Join(texts, ""), Model: out.Model}, nil
}

// Status implements provider.Adapter via GET /v1/models.
func (a *Adapter) Status(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+"/v1/models", nil)
	if err != nil {
		return err
	}
	a.headers(httpReq)
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("anthropic: unexpected status %s", resp.Status)
	}
	return nil
}
