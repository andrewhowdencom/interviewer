package http

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"
)

// NewClient creates a new HTTP client with an optional custom DNS server.
// If dnsServer is empty, the system's default DNS resolver is used.
func NewClient(dnsServer string) *http.Client {
	if dnsServer == "" {
		return http.DefaultClient
	}

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			// If the dnsServer is an IPv6 address, it needs to be wrapped in square brackets.
			addr := dnsServer + ":53"
			if strings.Contains(dnsServer, ":") {
				addr = "[" + dnsServer + "]:53"
			}
			return d.DialContext(ctx, "udp", addr)
		},
	}

	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Resolver:  resolver,
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &http.Client{
		Transport: transport,
	}
}
