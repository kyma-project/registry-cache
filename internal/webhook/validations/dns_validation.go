package validations

import (
	"context"
	"net"
	"time"
)

// DNSValidation defines DNS resolvability checks used by the validator.
//
//go:generate mockery --name=DNSValidator
type DNSValidator interface {
	// IsResolvable returns true if the given host has at least one A/AAAA record.
	IsResolvable(host string) bool
}

// DefaultDNSValidator implements DNSValidator using the net.Resolver to check host resolvability.
type DefaultDNSValidator struct {
	// optional timeout per lookup; defaults to 2s if zero
	Timeout time.Duration
}

func (r DefaultDNSValidator) IsResolvable(host string) bool {
	if host == "" {
		return false
	}
	res := net.DefaultResolver

	timeout := r.Timeout
	if timeout == 0 {
		timeout = 2 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Try lookup; any result counts as resolvable.
	if addrs, err := res.LookupHost(ctx, host); err == nil && len(addrs) > 0 {
		return true
	}
	return false
}
