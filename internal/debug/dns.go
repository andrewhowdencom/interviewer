package debug

import (
	"log/slog"
	"os"
)

// LogDNSInfo logs the contents of /etc/resolv.conf and /etc/hosts.
func LogDNSInfo() {
	logFile(
		"/etc/resolv.conf",
		"could not read /etc/resolv.conf",
	)
	logFile(
		"/etc/hosts",
		"could not read /etc/hosts",
	)
}

func logFile(path, errMsg string) {
	content, err := os.ReadFile(path)
	if err != nil {
		slog.Debug(errMsg, "error", err)
		return
	}
	slog.Debug("read file", "path", path, "content", string(content))
}
