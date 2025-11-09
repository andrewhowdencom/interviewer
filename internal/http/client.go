package http

import (
	"context"
	"fmt"
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

// NewIPv4Client creates a new HTTP client that forces connections to use IPv4.
func NewIPv4Client() *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Separate host and port.
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				// If splitting fails, it might be a host without a port.
				// Try to resolve it directly.
				host = addr
			}

			// Resolve the hostname to IPv4 addresses only.
			addrs, err := net.DefaultResolver.LookupIP(ctx, "ip4", host)
			if err != nil {
				return nil, err
			}
			if len(addrs) == 0 {
				return nil, fmt.Errorf("no IPv4 addresses found for %s", host)
			}

			// Dial the first resolved IPv4 address.
			dialer := &net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}

			var dialAddr string
			if port != "" {
				dialAddr = net.JoinHostPort(addrs[0].String(), port)
			} else {
				dialAddr = addrs[0].String()
			}


			return dialer.DialContext(ctx, network, dialAddr)
		},
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
