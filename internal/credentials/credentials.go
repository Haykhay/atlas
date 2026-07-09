// Package credentials stores provider API keys in the operating
// system's secure credential store. Secrets never touch the config
// file or logs.
package credentials

import (
	"errors"

	"github.com/zalando/go-keyring"
)

const service = "atlas"

// Set stores the API key for a provider.
func Set(provider, key string) error {
	return keyring.Set(service, provider, key)
}

// Get returns the stored API key for a provider.
func Get(provider string) (string, error) {
	return keyring.Get(service, provider)
}

// Delete removes the stored API key for a provider.
func Delete(provider string) error {
	return keyring.Delete(service, provider)
}

// IsNotFound reports whether err means no credential is stored.
func IsNotFound(err error) bool {
	return errors.Is(err, keyring.ErrNotFound)
}

// MockForTesting swaps the OS keychain for an in-memory store.
func MockForTesting() {
	keyring.MockInit()
}
