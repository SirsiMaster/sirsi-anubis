package mobile

import (
	"context"
	"time"

	"github.com/SirsiMaster/sirsi-pantheon/internal/ka"
)

// KaHunt scans for ghost app residuals and returns findings as JSON.
// Returns Response JSON with []Ghost data.
func KaHunt(includeSudo bool) string {
	scanner := ka.NewScanner()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	ghosts, err := scanner.Scan(ctx, includeSudo)
	if err != nil {
		return errorJSON("ghost scan failed: " + err.Error())
	}

	return successJSON(ghosts)
}

// KaEnumerateApps lists all installed applications with ghost status.
// Returns Response JSON with []InstalledApp data.
func KaEnumerateApps() string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	apps, err := ka.EnumerateApps(ctx)
	if err != nil {
		return errorJSON("app enumeration failed: " + err.Error())
	}

	return successJSON(apps)
}
