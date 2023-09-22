// Package resolve is responsible for everything DNS-related in gocurl.
package resolve

import (
	"fmt"
	"net"

	"github.com/AdguardTeam/dnsproxy/upstream"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/netutil/sysresolv"
	ctls "github.com/ameshkov/cfcrypto/tls"
	"github.com/ameshkov/gocurl/internal/config"
	"github.com/ameshkov/gocurl/internal/output"
	"github.com/miekg/dns"
)

// ErrEmptyResponse means that the response does not contain necessary RRs.
const ErrEmptyResponse = errors.Error("empty response")

// ErrNoResolvers means that system resolvers couldn't be discovered.
const ErrNoResolvers = errors.Error("no resolvers")

// ErrInvalidResolver means that the configured resolver is invalid.
const ErrInvalidResolver = errors.Error("invalid resolver")

// Resolver is a structure that is used whenever DNS resolution is required.
//
// TODO(ameshkov): Add --resolve parameter support.
type Resolver struct {
	cfg *config.Config
	out *output.Output

	// upstreams is the list of system resolvers to use.
	upstreams []upstream.Upstream
}

// NewResolver creates a new instance of *Resolver.
func NewResolver(cfg *config.Config, out *output.Output) (r *Resolver, err error) {
	var upstreams []upstream.Upstream

	if len(cfg.DNSServers) > 0 {
		out.Debug("Using custom configured DNS servers")
		upstreams = cfg.DNSServers
	} else {
		upstreams, err = getSystemResolvers()
		if err != nil {
			return nil, err
		}
	}

	return &Resolver{
		cfg:       cfg,
		out:       out,
		upstreams: upstreams,
	}, nil
}

// LookupHost looks up all IP addresses of the hostname.
func (r *Resolver) LookupHost(hostname string) (ipAddresses []net.IP, err error) {
	r.out.Debug("Resolving IP addresses of %s", hostname)

	ip := net.ParseIP(hostname)
	if ip != nil {
		// Trim zero bytes.
		if ip.To4() != nil {
			ip = ip.To4()
		}

		ipAddresses = append(ipAddresses, ip)

		return ipAddresses, nil
	}

	if addrs, ok := r.lookupFromCfg(hostname); ok {
		r.out.Debug("Resolved IP addresses for %s from the configuration", hostname)

		return addrs, nil
	}

	var errs []error

	for _, qType := range []uint16{dns.TypeA, dns.TypeAAAA} {
		msg := newMsg(hostname, qType)

		resp, u, dnsErr := dnsLookupAll(msg, r.upstreams)
		if dnsErr != nil {
			errs = append(errs, dnsErr)

			// try another qType now.
			continue
		}

		for _, rr := range resp.Answer {
			switch v := rr.(type) {
			case *dns.A:
				ipAddresses = append(ipAddresses, v.A)
			case *dns.AAAA:
				ipAddresses = append(ipAddresses, v.AAAA)
			}
		}

		r.out.Debug("%s responses received from %s", dns.Type(qType), u.Address())
	}

	if len(ipAddresses) == 0 {
		return nil, errors.Join(ErrEmptyResponse, errors.Join(errs...))
	}

	r.out.Debug("Found the following IP addresses for %s", hostname)
	for _, ipAddr := range ipAddresses {
		r.out.Debug("IP: %s", ipAddr)
	}

	return ipAddresses, nil
}

// LookupECHConfigs attempts to discover ECH configurations in DNS records of
// the specified hostname.  If no ECH configuration can be discovered for this
// domain, the function returns ErrEmptyResponse (checked via errors.Is/As).
func (r *Resolver) LookupECHConfigs(hostname string) (echConfigs []ctls.ECHConfig, err error) {
	r.out.Debug("Resolving ECH configuration for %s", hostname)

	if len(r.cfg.ECHConfigs) > 0 {
		r.out.Debug("Return pre-configured ECH configuration for %s", hostname)

		return r.cfg.ECHConfigs, nil
	}

	m := newMsg(hostname, dns.TypeHTTPS)

	var resp *dns.Msg
	var u upstream.Upstream
	resp, u, err = dnsLookupAll(m, r.upstreams)
	if err != nil {
		return nil, err
	}

	r.out.Debug("ECH configuration resolved using %s", u.Address())

	// Find all ECH configurations in the HTTPS records.
	var errs []error

	for _, rr := range resp.Answer {
		switch v := rr.(type) {
		case *dns.HTTPS:
			for _, svcb := range v.SVCB.Value {
				if svcb.Key() == dns.SVCB_ECHCONFIG {
					echConfigRR := svcb.(*dns.SVCBECHConfig)
					echConfig, echErr := ctls.UnmarshalECHConfigs(echConfigRR.ECH)
					if echErr != nil {
						r.out.Debug("Invalid ECH configuration: %v", echErr)
						errs = append(errs, echErr)
					} else {
						echConfigs = append(echConfigs, echConfig...)
					}
				}
			}
		}
	}

	if len(echConfigs) == 0 {
		return nil, errors.Join(ErrEmptyResponse, errors.Join(errs...))
	}

	return echConfigs, nil
}

// lookupFromCfg checks if IP address for hostname are specified in the
// configuration.
func (r *Resolver) lookupFromCfg(hostname string) (addrs []net.IP, ok bool) {
	if len(r.cfg.Resolve) == 0 {
		return nil, false
	}

	if addrs, ok = r.cfg.Resolve[hostname]; ok {
		return addrs, ok
	}

	if addrs, ok = r.cfg.Resolve["*"]; ok {
		return addrs, ok
	}

	return nil, false
}

// dnsLookupAll sends the query m to each DNS resolver until it gets
// a successful non-empty response.  If all attempts are unsuccessful, returns
// an error.
func dnsLookupAll(m *dns.Msg, upstreams []upstream.Upstream) (resp *dns.Msg, u upstream.Upstream, err error) {
	var errs []error

	for _, u = range upstreams {
		var dnsErr error
		resp, dnsErr = dnsLookup(m, u)
		if dnsErr != nil {
			errs = append(errs, dnsErr)
		} else {
			return resp, u, nil
		}
	}

	return nil, nil, errors.List("dns lookup", errs...)
}

// dnsLookup sends the query m over to DNS resolver addr and returns the
// response.  Adds additional logic on top of it: returns an error when the
// response code is not success or when there are no resource records.
func dnsLookup(m *dns.Msg, u upstream.Upstream) (resp *dns.Msg, err error) {
	resp, err = u.Exchange(m)
	qTypeStr := dns.Type(m.Question[0].Qtype).String()

	if err != nil {
		return nil, err
	}

	if resp.Rcode != dns.RcodeSuccess {
		return nil, fmt.Errorf(
			"dns response %s code from %s: %s",
			qTypeStr,
			u.Address(),
			rCodeToString(resp.Rcode),
		)
	}

	if len(resp.Answer) == 0 {
		return nil,
			errors.Annotate(ErrEmptyResponse, "no %s resource records from %s: %w", qTypeStr, u.Address())
	}

	return resp, nil
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

// getSystemResolvers returns a list of upstream.Upstream that were created
// from system resolvers.
func getSystemResolvers() (upstreams []upstream.Upstream, err error) {
	sr, err := sysresolv.NewSystemResolvers(nil)
	if err != nil {
		return nil, err
	}

	addrs := sr.Addrs()
	for _, addr := range addrs {
		u, uErr := upstream.AddressToUpstream(addr, nil)
		if uErr != nil {
			return nil, errors.Join(ErrInvalidResolver, uErr)
		}

		upstreams = append(upstreams, u)
	}

	if len(upstreams) == 0 {
		return nil, ErrNoResolvers
	}

	return upstreams, nil
}
