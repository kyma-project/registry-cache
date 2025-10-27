package validations

import (
	"context"
	"net"
	"time"
)

// DNSValidator defines DNS resolvability checks used by the validator.
//
//go:generate mockery --name=DNSValidator
type DNSValidator interface {
	// IsResolvable returns true if the given host has at least one A/AAAA record.
	IsResolvable(host string) bool
}

// DefaultDNSValidator implements DNSValidator using the net.Resolver to check host resolvability.
type DefaultDNSValidator struct{}

func (r DefaultDNSValidator) IsResolvable(host string) bool {
	if host == "" {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if addrs, err := net.DefaultResolver.LookupHost(ctx, host); err == nil && len(addrs) > 0 {
		return true
	}
	return false
}
