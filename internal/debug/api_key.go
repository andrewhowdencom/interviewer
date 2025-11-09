package debug

import (
	"log/slog"
)

// LogGeminiAPIKey logs the last 4 characters of the Gemini API key.
func LogGeminiAPIKey(apiKey string) {
	if len(apiKey) > 4 {
		slog.Debug("using gemini api key", "suffix", apiKey[len(apiKey)-4:])
	}
}
