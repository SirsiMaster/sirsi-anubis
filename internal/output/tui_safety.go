package output

import (
	"sync"

	"github.com/SirsiMaster/sirsi-pantheon/internal/jackal"
)

// SafetyGateway centralizes all destructive action confirmation.
// All cleanup flows must pass through this gateway before deleting files.
type SafetyGateway interface {
	ConfirmClean(items []jackal.Finding, source string) error
}

// defaultSafetyGateway preserves current behavior — always allows cleanup.
type defaultSafetyGateway struct{}

func (g *defaultSafetyGateway) ConfirmClean(_ []jackal.Finding, _ string) error {
	return nil // current behavior: no additional gate beyond user selection
}

// Package-level gateway (Rule A21: concurrency-safe injectable mocks).
var (
	cleanGatewayMu sync.RWMutex
	cleanGateway   SafetyGateway = &defaultSafetyGateway{}
)

func getCleanGateway() SafetyGateway {
	cleanGatewayMu.RLock()
	defer cleanGatewayMu.RUnlock()
	return cleanGateway
}

func setCleanGateway(gw SafetyGateway) {
	cleanGatewayMu.Lock()
	defer cleanGatewayMu.Unlock()
	cleanGateway = gw
}
