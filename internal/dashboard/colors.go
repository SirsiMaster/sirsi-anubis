// Package dashboard serves the Pantheon local dashboard — a self-contained
// HTML UI at localhost:9119. Menubar clicks open browser pages; CLI stays
// in the terminal. Both surfaces read from the same data stores.
package dashboard

// Brand colors for HTML templates (mirrors internal/output/terminal.go).
const (
	ColorGold     = "#C8A951"
	ColorBlack    = "#0F0F0F"
	ColorLapis    = "#1A1A5E"
	ColorWhite    = "#FAFAFA"
	ColorDim      = "#888888"
	ColorRed      = "#FF4444"
	ColorGreen    = "#44FF88"
	ColorYellow   = "#FFD700"
	ColorBg       = "#06060F"
	ColorBgPanel  = "rgba(6,6,15,.88)"
	ColorBorder   = "rgba(200,169,81,.12)"
	ColorBorderHi = "rgba(200,169,81,.35)"
)

// DashboardPort is the fixed port for the dashboard server.
const DashboardPort = 9119
