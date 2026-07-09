package cli

import (
	"errors"

	"github.com/Haykhay/atlas/internal/config"
	"github.com/Haykhay/atlas/internal/credentials"
	"github.com/Haykhay/atlas/internal/provider"
)

// defaultAdapter builds the user's default provider adapter, pulling
// the API key from the OS keychain. A missing credential is tolerated
// (local providers like ollama need none).
func defaultAdapter(cfg *config.Config) (provider.Adapter, error) {
	name := cfg.DefaultProvider
	if name == "" {
		return nil, errors.New("no default provider configured (run: atlas configure)")
	}
	key, err := credentials.Get(name)
	if err != nil && !credentials.IsNotFound(err) {
		return nil, err
	}
	pc := cfg.Providers[name]
	return provider.New(name, provider.Settings{APIKey: key, BaseURL: pc.BaseURL, Model: pc.Model})
}
