package mobile

import (
	"encoding/json"

	"github.com/SirsiMaster/sirsi-pantheon/internal/thoth"
)

// ThothInit initializes a .thoth/ knowledge system in the given project root.
// Returns Response JSON with ProjectInfo data.
func ThothInit(projectRoot string) string {
	opts := thoth.InitOptions{
		RepoRoot: projectRoot,
		Yes:      true, // Non-interactive on mobile
	}

	if err := thoth.Init(opts); err != nil {
		return errorJSON("thoth init failed: " + err.Error())
	}

	info := thoth.DetectProject(projectRoot)
	return successJSON(info)
}

// ThothSync synchronizes project memory with source facts.
// optionsJSON accepts: {"root": "/path", "verbose": true}
func ThothSync(optionsJSON string) string {
	var opts thoth.SyncOptions
	if optionsJSON != "" {
		if err := json.Unmarshal([]byte(optionsJSON), &opts); err != nil {
			return errorJSON("invalid options: " + err.Error())
		}
	}

	if err := thoth.Sync(opts); err != nil {
		return errorJSON("thoth sync failed: " + err.Error())
	}

	return successJSON(map[string]string{"status": "synced"})
}

// ThothCompact compresses project memory for context efficiency.
// optionsJSON accepts: {"root": "/path"}
func ThothCompact(optionsJSON string) string {
	var opts thoth.CompactOptions
	if optionsJSON != "" {
		if err := json.Unmarshal([]byte(optionsJSON), &opts); err != nil {
			return errorJSON("invalid options: " + err.Error())
		}
	}

	if err := thoth.Compact(opts); err != nil {
		return errorJSON("thoth compact failed: " + err.Error())
	}

	return successJSON(map[string]string{"status": "compacted"})
}

// ThothDetectProject returns auto-detected project metadata.
// Returns Response JSON with ProjectInfo data.
func ThothDetectProject(projectRoot string) string {
	info := thoth.DetectProject(projectRoot)
	return successJSON(info)
}
