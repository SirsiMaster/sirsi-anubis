// Package mobile provides the gomobile binding layer for Pantheon iOS.
// All exported functions use JSON string I/O for cross-language compatibility.
// Build: gomobile bind -target=ios -o PantheonCore.xcframework ./mobile/
package mobile

import "encoding/json"

// Version returns the Pantheon mobile SDK version.
func Version() string {
	return "0.17.0"
}

// Response is the standard envelope for all mobile bridge responses.
type Response struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data,omitempty"`
	Error string          `json:"error,omitempty"`
}

func successJSON(data any) string {
	raw, err := json.Marshal(data)
	if err != nil {
		return errorJSON("marshal: " + err.Error())
	}
	resp := Response{OK: true, Data: raw}
	out, _ := json.Marshal(resp)
	return string(out)
}

func errorJSON(msg string) string {
	resp := Response{OK: false, Error: msg}
	out, _ := json.Marshal(resp)
	return string(out)
}
