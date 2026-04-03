package version

// ModuleVersion tracks the version of each deity module.
var Modules = map[string]string{
	"ka":      "1.1.0", // v1.1.0: multi-layer ghost matching, app enumerator, uninstaller
	"anubis":  "1.0.0", // scanning, cleaning, safety
	"thoth":   "1.0.0", // memory, sync, compact
	"maat":    "1.0.0", // audit, heal, coverage
	"seshat":  "2.0.0", // v2: universal knowledge grafting, 5+3 adapters, Chrome profiles
	"hapi":    "1.0.0", // hardware detection, accelerator routing
	"seba":    "1.0.0", // topology mapping, diagrams
	"horus":   "1.0.0", // filesystem index, sight
	"sekhmet": "1.0.0", // watchdog, process guard
	"khepri":  "1.0.0", // network scan, container audit
	"isis":    "1.0.0", // remediation engine
	"neith":   "1.0.0", // loom, scope assembly, drift detection
	"ra":      "1.0.0", // orchestrator, deploy, pipeline
	"osiris":  "0.5.0", // FinalWishes checkpoint (partial)
	"hathor":  "1.0.0", // mirror dedup
}

// Get returns the version of a module, or "unknown" if not registered.
func Get(module string) string {
	if v, ok := Modules[module]; ok {
		return v
	}
	return "unknown"
}
