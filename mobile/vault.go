package mobile

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/SirsiMaster/sirsi-pantheon/internal/vault"
)

// vaultOnce provides lazy-init singleton for the Vault Store.
// Mobile apps cannot pass pointers across gomobile, so we manage
// a single Store instance at the package level.
var (
	vaultOnce  sync.Once
	vaultStore *vault.Store
	vaultErr   error
)

// vaultPath returns the mobile-appropriate vault database path.
// On mobile, this lives under the app's documents directory.
func vaultPath() string {
	// Use PANTHEON_VAULT_PATH env var if set (for testing).
	if p := os.Getenv("PANTHEON_VAULT_PATH"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "sirsi", "vault", "context.db")
}

// getVault returns the singleton vault store, initializing it lazily.
func getVault() (*vault.Store, error) {
	vaultOnce.Do(func() {
		vaultStore, vaultErr = vault.Open(vaultPath())
	})
	return vaultStore, vaultErr
}

// resetVault resets the singleton for testing purposes.
func resetVault() {
	if vaultStore != nil {
		vaultStore.Close()
	}
	vaultOnce = sync.Once{}
	vaultStore = nil
	vaultErr = nil
}

// VaultStore sandboxes a piece of output into the context vault.
// Returns Response JSON with the stored Entry data.
func VaultStore(source, tag, content string, tokens int) string {
	store, err := getVault()
	if err != nil {
		return errorJSON("vault open: " + err.Error())
	}

	entry, err := store.Store(source, tag, content, tokens)
	if err != nil {
		return errorJSON("vault store: " + err.Error())
	}

	return successJSON(entry)
}

// VaultSearch performs FTS5 full-text search in the context vault.
// Returns Response JSON with SearchResult data.
func VaultSearch(query string, limit int) string {
	store, err := getVault()
	if err != nil {
		return errorJSON("vault open: " + err.Error())
	}

	if query == "" {
		return errorJSON("query is required")
	}

	result, err := store.Search(query, limit)
	if err != nil {
		return errorJSON("vault search: " + err.Error())
	}

	return successJSON(result)
}

// VaultGet retrieves a specific vault entry by ID with full content.
// Returns Response JSON with Entry data.
func VaultGet(id int64) string {
	store, err := getVault()
	if err != nil {
		return errorJSON("vault open: " + err.Error())
	}

	entry, err := store.Get(id)
	if err != nil {
		return errorJSON(fmt.Sprintf("vault get %d: %s", id, err.Error()))
	}

	return successJSON(entry)
}

// VaultStats returns vault statistics (entry count, total bytes, tokens, tag breakdown).
// Returns Response JSON with StoreStats data.
func VaultStats() string {
	store, err := getVault()
	if err != nil {
		return errorJSON("vault open: " + err.Error())
	}

	stats, err := store.Stats()
	if err != nil {
		return errorJSON("vault stats: " + err.Error())
	}

	return successJSON(stats)
}

// VaultPrune removes vault entries older than the given number of hours.
// Returns Response JSON with {"pruned": N} count.
func VaultPrune(olderThanHours int) string {
	store, err := getVault()
	if err != nil {
		return errorJSON("vault open: " + err.Error())
	}

	if olderThanHours <= 0 {
		return errorJSON("olderThanHours must be positive")
	}

	dur := time.Duration(olderThanHours) * time.Hour
	count, err := store.Prune(dur)
	if err != nil {
		return errorJSON("vault prune: " + err.Error())
	}

	return successJSON(map[string]int{"pruned": count})
}
