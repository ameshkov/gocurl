// Package resolve is responsible for everything DNS-related in gocurl.
package resolve

import (
	"fmt"
	"net"

	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/netutil/sysresolv"
	"github.com/ameshkov/gocurl/internal/config"
	"github.com/ameshkov/gocurl/internal/output"
	"github.com/miekg/dns"
)

// Resolver is a structure that is used whenever DNS resolution is required.
//
// TODO(ameshkov): Add --resolve parameter support.
type Resolver struct {
	cfg *config.Config
	out *output.Output

	// addrs is the list of system resolvers to use.
	addrs []string
}

// NewResolver creates a new instance of *Resolver.
func NewResolver(cfg *config.Config, out *output.Output) (r *Resolver, err error) {
	sr, err := sysresolv.NewSystemResolvers(nil)
	if err != nil {
		return nil, err
	}

	addrs := sr.Get()
	if len(addrs) == 0 {
		return nil, errors.Error("resolve: no resolvers found")
	}

	return &Resolver{
		cfg:   cfg,
		out:   out,
		addrs: addrs,
	}, nil
}

// LookupHost looks up all IP addresses of the hostname.
func (r *Resolver) LookupHost(hostname string) (ipAddresses []net.IP, err error) {
	// TODO: logging

	ip := net.ParseIP(hostname)
	if ip != nil {
		// Trim zero bytes.
		if ip.To4() != nil {
			ip = ip.To4()
		}

		ipAddresses = append(ipAddresses, ip)
		return ipAddresses, nil
	}

	var errs []error

	for _, qType := range []uint16{dns.TypeA, dns.TypeAAAA} {
		msg := newMsg(hostname, qType)

		for _, addr := range r.addrs {
			respIPs, dnsErr := lookupIPAddresses(msg, addr)
			if dnsErr != nil {
				errs = append(errs, dnsErr)
			} else {
				ipAddresses = append(ipAddresses, respIPs...)

				// If the IP addresses for the qType were retrieved successfully,
				// break the inner cycle and get to the next query type.
				break
			}
		}
	}

	if len(ipAddresses) == 0 {
		return nil, errors.List("failed to lookup", errs...)
	}

	return ipAddresses, nil
}

// lookupIPAddresses makes a DNS query to the specified address and returns all
// IP addresses from the response.
func lookupIPAddresses(m *dns.Msg, addr string) (ipAddresses []net.IP, err error) {
	resp, err := dns.Exchange(m, net.JoinHostPort(addr, "53"))
	if err != nil {
		return nil, err
	}

	if resp.Rcode != dns.RcodeSuccess {
		return nil, fmt.Errorf("dns response code from %s: %s", addr, rCodeToString(resp.Rcode))
	}

	for _, rr := range resp.Answer {
		switch v := rr.(type) {
		case *dns.A:
			ipAddresses = append(ipAddresses, v.A)
		case *dns.AAAA:
			ipAddresses = append(ipAddresses, v.AAAA)
		}
	}

	if len(ipAddresses) == 0 {
		return nil, fmt.Errorf("no IP addresses in response from %s", addr)
	}

	return ipAddresses, nil
}

// newMsg creates new *dns.Msg of the specified type for hostname.
func newMsg(hostname string, qType uint16) (m *dns.Msg) {
	m = &dns.Msg{}
	m.Id = dns.Id()
	m.RecursionDesired = true
	m.Question = []dns.Question{{
		Name:   dns.Fqdn(hostname),
		Qtype:  qType,
		Qclass: dns.ClassINET,
	}}

	return m
}

// rCodeToString is a helper function to convert DNS message response code to
// string.
func rCodeToString(rCode int) (str string) {
	if v, ok := dns.RcodeToString[rCode]; ok {
		return v
	}

	return fmt.Sprintf("TYPE_%d", rCode)
}
