package eref

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/gotomicro/ego/core/elog"
	"net"
	"net/http"
	"regexp"
	"strings"
)

var (
	// De-facto standard header keys.
	xForwardedFor    = http.CanonicalHeaderKey("X-Forwarded-For")
	xForwardedHost   = http.CanonicalHeaderKey("X-Forwarded-Host")
	xForwardedProto  = http.CanonicalHeaderKey("X-Forwarded-Proto")
	xForwardedScheme = http.CanonicalHeaderKey("X-Forwarded-Scheme")
	xRealIP          = http.CanonicalHeaderKey("X-Real-IP")
)

var (
	// RFC7239 defines a new "Forwarded: " header designed to replace the
	// existing use of X-Forwarded-* headers.
	// e.g. Forwarded: for=192.0.2.60;proto=https;by=203.0.113.43
	forwarded = http.CanonicalHeaderKey("Forwarded")
	// Allows for a sub-match of the first value after 'for=' to the next
	// comma, semi-colon or space. The match is case-insensitive.
	forRegex = regexp.MustCompile(`(?i)(?:for=)([^(;|,| )]+)`)
	// Allows for a sub-match for the first instance of scheme (http|https)
	// prefixed by 'proto='. The match is case-insensitive.
	protoRegex = regexp.MustCompile(`(?i)(?:proto=)(https|http)`)
)

func ProxyIpMiddleware(logger *elog.Component, config *Config) restful.FilterFunction {
	return Filter(func(ctx FilterContext) {
		// Set the remote IP with the value passed from the proxy.
		ip := getIP(ctx.Req())
		// IP 写入上下文
		ctx.SetAttribute("ip", ip)
		// Set the scheme (proto) with the value passed from the proxy.
		if scheme := getScheme(ctx.Req()); scheme != "" {
			ctx.Req().URL.Scheme = scheme
		}
		// Set the host with the value passed by the proxy
		if ctx.HeaderParameter(xForwardedHost) != "" {
			ctx.Req().Host = ctx.HeaderParameter(xForwardedHost)
		}
		ctx.ProcessFilter()
	})
}

// getIP retrieves the IP from the X-Forwarded-For, X-Real-IP and RFC7239
// Forwarded headers (in that order).
func getIP(r *http.Request) string {
	var fwd string
	var err error

	addr := strings.TrimSpace(r.RemoteAddr)
	if fwd = r.Header.Get(xForwardedFor); fwd != "" {
		// Only grab the first (client) address. Note that '192.168.0.1,
		// 10.1.1.1' is a valid key for X-Forwarded-For where addresses after
		// the first may represent forwarding proxies earlier in the chain.
		s := strings.Index(fwd, ", ")
		if s == -1 {
			s = len(fwd)
		}
		addr = fwd[:s]
	} else if fwd = r.Header.Get(xRealIP); fwd != "" {
		// X-Real-IP should only contain one IP address (the client making the
		// request).
		addr = fwd
	} else if fwd = r.Header.Get(forwarded); fwd != "" {
		// match should contain at least two elements if the protocol was
		// specified in the Forwarded header. The first element will always be
		// the 'for=' capture, which we ignore. In the case of multiple IP
		// addresses (for=8.8.8.8, 8.8.4.4,172.16.1.20 is valid) we only
		// extract the first, which should be the client IP.
		if match := forRegex.FindStringSubmatch(fwd); len(match) > 1 {
			// IPv6 addresses in Forwarded headers are quoted-strings. We strip
			// these quotes.
			addr = strings.Trim(match[1], `"`)
		}
	}

	addr, _, err = net.SplitHostPort(addr)
	if err != nil {
		return ""
	}
	return addr
}

// getScheme retrieves the scheme from the X-Forwarded-Proto and RFC7239
// Forwarded headers (in that order).
func getScheme(r *http.Request) string {
	var scheme string

	// Retrieve the scheme from X-Forwarded-Proto.
	if proto := r.Header.Get(xForwardedProto); proto != "" {
		scheme = strings.ToLower(proto)
	} else if proto = r.Header.Get(xForwardedScheme); proto != "" {
		scheme = strings.ToLower(proto)
	} else if proto = r.Header.Get(forwarded); proto != "" {
		// match should contain at least two elements if the protocol was
		// specified in the Forwarded header. The first element will always be
		// the 'proto=' capture, which we ignore. In the case of multiple proto
		// parameters (invalid) we only extract the first.
		if match := protoRegex.FindStringSubmatch(proto); len(match) > 1 {
			scheme = strings.ToLower(match[1])
		}
	}

	return scheme
}
