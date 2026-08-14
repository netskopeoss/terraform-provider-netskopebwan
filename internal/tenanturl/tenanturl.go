// Package tenanturl addresses a tenant by rewriting the tenant domain of a URL.
//
// A Netskope hostname carries the tenant in its leading label, separated from
// the rest either by a dot or, where a dot would cost a wildcard certificate a
// level, by "-dot-": both "foo.api.example.net" and "foo-dot-api.example.net"
// name the tenant "foo". Swapping that label is how one URL reaches another
// tenant.
//
// This is a trimmed copy of the shared tenanturl package, kept to the functions
// the provider uses and otherwise identical to it, so the two can be diffed.
package tenanturl

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// tenantDomainIndex reports where the tenant domain of a hostname ends, and
// which of the two separators ends it.
func tenantDomainIndex(s string) (int, string) {
	if s == "" {
		return len(s), ""
	}

	letterSepIdx := strings.Index(s, "-dot-")
	symbolSepIdx := strings.Index(s, ".")

	i, sep := len(s), ""

	switch {
	case symbolSepIdx >= 0 && letterSepIdx >= 0:
		if letterSepIdx < symbolSepIdx {
			i = letterSepIdx
			sep = "-dot-"
		} else {
			i = symbolSepIdx
			sep = "."
		}
	case symbolSepIdx >= 0:
		i = symbolSepIdx
		sep = "."
	case letterSepIdx >= 0:
		i = letterSepIdx
		sep = "-dot-"
	}

	return i, sep
}

// ReplaceTenantDomainInHostname replaces the tenant domain of a hostname,
// keeping the separator and everything after it.
func ReplaceTenantDomainInHostname(s, repl string) (string, error) {
	i, _ := tenantDomainIndex(s)

	return repl + s[i:], nil
}

// ReplaceTenantDomainInHost is ReplaceTenantDomainInHostname for a host that may
// carry a port.
func ReplaceTenantDomainInHost(s, repl string) (string, error) {
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		return ReplaceTenantDomainInHostname(s, repl)
	}

	host, err = ReplaceTenantDomainInHostname(host, repl)
	if err != nil {
		return "", err
	}

	return net.JoinHostPort(host, port), nil
}

// ReplaceTenantDomainInURL points a URL at another tenant, leaving its scheme,
// port, path and query alone.
func ReplaceTenantDomainInURL(s, repl string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("parse url=%s: %w", s, err)
	}

	u.Host, err = ReplaceTenantDomainInHost(u.Host, repl)
	if err != nil {
		return "", err
	}

	return u.String(), nil
}
