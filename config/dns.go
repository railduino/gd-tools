package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/railduino/gd-tools/dns"
	"github.com/railduino/gd-tools/email"
)

func (cfg *Config) DNSProvider() (dns.DNSProvider, error) {
	if cfg.SkipDNS {
		return dns.NewNoopProvider("SkipDNS")
	}

	if cfg.HetznerToken != "" {
		return dns.NewHetznerProvider(cfg.HetznerToken)
	}
	if cfg.IonosToken != "" {
		return dns.NewIonosProvider(cfg.IonosToken)
	}
	// TODO (later) add more providers here

	return dns.NewNoopProvider("NoDNS")
}

func (cfg *Config) SetA(zone, name, ip string) (string, error) {
	record := dns.RRecord{
		Prio:  0,
		Value: ip,
	}
	records := []dns.RRecord{record}

	msg := fmt.Sprintf("SetA %s -> %s = %v (ttl=%d)\n", zone, name, records, cfg.RegTTL)
	if cfg.Verbose {
		cfg.Debug(msg)
	} else if cfg.Dry {
		cfg.Say(msg)
		return "", nil
	}

	p, err := cfg.DNSProvider()
	if err != nil {
		return "", err
	}
	if p == nil {
		return "", fmt.Errorf("missing DNS provider")
	}
	ctx := context.Background()

	return p.UpsertRRSet(ctx, zone, dns.RRSet{
		Name:    name, // relative: "www", "autodiscover", ...
		Type:    dns.RR_A,
		TTL:     cfg.RegTTL,
		Records: records,
	})
}

func (cfg *Config) SetAAAA(zone, name, ip string) (string, error) {
	record := dns.RRecord{
		Prio:  0,
		Value: ip,
	}
	records := []dns.RRecord{record}

	msg := fmt.Sprintf("SetAAAA %s -> %s = %v (ttl=%d)\n", zone, name, records, cfg.RegTTL)
	if cfg.Verbose {
		cfg.Debug(msg)
	} else if cfg.Dry {
		cfg.Say(msg)
		return "", nil
	}

	p, err := cfg.DNSProvider()
	if err != nil {
		return "", err
	}
	if p == nil {
		return "", fmt.Errorf("missing DNS provider")
	}
	ctx := context.Background()

	return p.UpsertRRSet(ctx, zone, dns.RRSet{
		Name:    name, // relative: "www", "autodiscover", ...
		Type:    dns.RR_AAAA,
		TTL:     cfg.RegTTL,
		Records: records,
	})
}

// SetHostSPF sets an explicit SPF record for an SMTP host.
// This is required by modern receivers (Microsoft, Apple, Google)
// which validate SPF against the envelope/HELO domain, not only the From domain.
func (cfg *Config) SetHostSPF(zone, name, ip4, ip6 string) (string, error) {
	record := dns.RRecord{
		Prio:  0,
		Value: fmt.Sprintf("v=spf1 ip4:%s ip6:%s -all", ip4, ip6),
	}
	records := []dns.RRecord{record}

	msg := fmt.Sprintf("SetHostSPF %s -> %s = %v (ttl=%d)\n", zone, name, records, cfg.RegTTL)
	if cfg.Verbose {
		cfg.Debug(msg)
	} else if cfg.Dry {
		cfg.Say(msg)
		return "", nil
	}

	p, err := cfg.DNSProvider()
	if err != nil {
		return "", err
	}
	if p == nil {
		return "", fmt.Errorf("missing DNS provider")
	}
	ctx := context.Background()

	return p.UpsertRRSet(ctx, zone, dns.RRSet{
		Name:    name, // relative: "www", "autodiscover", ...
		Type:    dns.RR_TXT,
		TTL:     cfg.RegTTL,
		Records: records,
	})
}

func (cfg *Config) SetCNAME(zone, name string) (string, error) {
	record := dns.RRecord{
		Prio:  0,
		Value: cfg.FQDN(),
	}
	records := []dns.RRecord{record}

	msg := fmt.Sprintf("SetCNAME %s -> %s = %v (ttl=%d)\n", zone, name, records, cfg.RegTTL)
	if cfg.Verbose {
		cfg.Debug(msg)
	} else if cfg.Dry {
		cfg.Say(msg)
		return "", nil
	}

	p, err := cfg.DNSProvider()
	if err != nil {
		return "", err
	}
	if p == nil {
		return "", fmt.Errorf("missing DNS provider")
	}
	ctx := context.Background()

	return p.UpsertRRSet(ctx, zone, dns.RRSet{
		Name:    name, // relative: "www", "autodiscover", ...
		Type:    dns.RR_CNAME,
		TTL:     cfg.RegTTL,
		Records: records,
	})
}

func (cfg *Config) UpdateDomainDNS(domain *email.Domain) ([]string, error) {
	var result []string

	p, err := cfg.DNSProvider()
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, fmt.Errorf("missing DNS provider")
	}
	ctx := context.Background()

	// SPF
	if status, err := cfg.SetApexTXT(ctx, p, domain); err != nil {
		return nil, fmt.Errorf("[dns] failed to set ApexTXT: %w", err)
	} else if status != "" {
		result = append(result, status)
	}

	// MTA-STS
	if status, err := cfg.SetSTS(ctx, p, domain); err != nil {
		return nil, fmt.Errorf("[dns] failed to set STS: %w", err)
	} else if status != "" {
		result = append(result, status)
	}

	// DKIM(s)
	for _, dkim := range domain.DKIMs {
		if status, err := cfg.SetDKIM(ctx, p, domain, dkim); err != nil {
			return nil, fmt.Errorf("[dns] failed to set DKIM: %w", err)
		} else if status != "" {
			result = append(result, status)
		}
	}

	// DMARC
	if status, err := cfg.SetDMARC(ctx, p, domain); err != nil {
		return nil, fmt.Errorf("[dns] failed to set DMARC: %w", err)
	} else if status != "" {
		result = append(result, status)
	}

	// MX
	if status, err := cfg.SetMX(ctx, p, domain); err != nil {
		return nil, fmt.Errorf("[dns] failed to set MX: %w", err)
	} else if status != "" {
		result = append(result, status)
	}

	// CAA
	if status, err := cfg.SetCAA(ctx, p, domain); err != nil {
		return nil, fmt.Errorf("[dns] failed to set CAA: %w", err)
	} else if status != "" {
		result = append(result, status)
	}

	return result, nil
}

func (cfg *Config) SetApexTXT(ctx context.Context, p dns.DNSProvider, domain *email.Domain) (string, error) {
	record := dns.RRecord{
		Prio:  0,
		Value: domain.GetSPF("ip4:"+cfg.IPv4Addr, "ip6:"+cfg.IPv6Addr),
	}
	records := []dns.RRecord{record}

	if domain.SpamBarrier != "" {
		records = append(records, dns.RRecord{
			Prio:  0,
			Value: domain.SpamBarrier,
		})
	}
	if domain.BrevoCode != "" {
		records = append(records, dns.RRecord{
			Prio:  0,
			Value: domain.BrevoCode,
		})
	}
	// TODO (later) add more top-level (Apex) records here, like:
	// Google-site-verification, apple-domain-verification, MS, etc.

	msg := fmt.Sprintf("SetApexTXT %s -> %v (ttl=%d)", domain.Name, records, cfg.RegTTL)
	if cfg.Verbose {
		cfg.Debug(msg)
	} else if cfg.Dry {
		cfg.Say(msg)
		return "", nil
	}

	return p.UpsertRRSet(ctx, domain.Name, dns.RRSet{
		Name:    "@",
		Type:    dns.RR_TXT,
		TTL:     cfg.RegTTL,
		Records: records,
	})
}

func (cfg *Config) SetSTS(ctx context.Context, p dns.DNSProvider, domain *email.Domain) (string, error) {
	record := dns.RRecord{
		Prio:  0,
		Value: "v=STSv1; id=20250724", // My birthday, in gd-tools first year :-)
	}
	records := []dns.RRecord{record}

	msg := fmt.Sprintf("SetSTS %s -> %v (ttl=%d)", domain.Name, records, cfg.RegTTL)
	if cfg.Verbose {
		cfg.Debug(msg)
	} else if cfg.Dry {
		cfg.Say(msg)
		return "", nil
	}

	return p.UpsertRRSet(ctx, domain.Name, dns.RRSet{
		Name:    "_mta-sts",
		Type:    dns.RR_TXT,
		TTL:     cfg.RegTTL,
		Records: records,
	})
}

func (cfg *Config) SetDKIM(ctx context.Context, p dns.DNSProvider, domain *email.Domain, dkim email.DKIM) (string, error) {
	selector := dkim.Selector + "._domainkey"

	// Decide record type
	var (
		rrType  dns.RRSetType
		rrValue string
	)

	if dkim.PubValue != "" {
		// Classic DKIM via TXT record (public key)
		rrType = dns.RR_TXT
		rrValue = "v=DKIM1; k=rsa; p=" + dkim.PubValue
	} else if dkim.CNAME != "" {
		// Provider-managed DKIM via CNAME delegation
		rrType = dns.RR_CNAME
		rrValue = normalizeFQDN(dkim.CNAME)
	} else {
		return "", fmt.Errorf("DKIM for %s: neither PubValue (TXT) nor CNAME target provided", domain.Name)
	}

	records := []dns.RRecord{{
		Prio:  0,
		Value: rrValue,
	}}

	preview := rrValue
	if rrType == dns.RR_TXT && len(preview) > 32 {
		preview = preview[:32] + "..."
	}

	msg := fmt.Sprintf("SetDKIM %s (%s) %s -> '%s' (ttl=%d)",
		domain.Name, selector, rrType, preview, cfg.RegTTL,
	)

	if cfg.Verbose {
		cfg.Debug(msg)
	} else if cfg.Dry {
		cfg.Say(msg)
		return "", nil
	}

	return p.UpsertRRSet(ctx, domain.Name, dns.RRSet{
		Name:    selector,
		Type:    rrType,
		TTL:     cfg.RegTTL,
		Records: records,
	})
}

func normalizeFQDN(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, ".")
	return s + "."
}

func (cfg *Config) SetDMARC(ctx context.Context, p dns.DNSProvider, domain *email.Domain) (string, error) {
	record := dns.RRecord{
		Prio:  0,
		Value: domain.DMARC,
	}
	records := []dns.RRecord{record}

	msg := fmt.Sprintf("SetDMARC %s (_dmarc) -> '%s' (ttl=%d)", domain.Name, domain.DMARC, cfg.RegTTL)
	if cfg.Verbose {
		cfg.Debug(msg)
	} else if cfg.Dry {
		cfg.Say(msg)
		return "", nil
	}

	return p.UpsertRRSet(ctx, domain.Name, dns.RRSet{
		Name:    "_dmarc",
		Type:    dns.RR_TXT,
		TTL:     cfg.RegTTL,
		Records: records,
	})
}

func (cfg *Config) SetMX(ctx context.Context, p dns.DNSProvider, domain *email.Domain) (string, error) {
	if cfg.SkipMX {
		return fmt.Sprintf("✅ DNS (mx) for %s skipped", domain.Name), nil
	}

	// Build "prio fqdn" list
	var records []dns.RRecord
	for _, mx := range domain.MXs {
		if mx.FQDN == "" || mx.Prio <= 0 {
			continue
		}
		record := dns.RRecord{
			Prio:  mx.Prio,
			Value: mx.FQDN,
		}
		records = append(records, record)
	}

	msg := fmt.Sprintf("SetMX %s -> %v (ttl=%d)", domain.Name, records, cfg.RegTTL)
	if cfg.Verbose {
		cfg.Debug(msg)
	} else if cfg.Dry {
		cfg.Say(msg)
		return "", nil
	}

	return p.UpsertRRSet(ctx, domain.Name, dns.RRSet{
		Name:    "@",
		Type:    dns.RR_MX,
		TTL:     cfg.RegTTL,
		Records: records,
	})
}

func (cfg *Config) SetCAA(ctx context.Context, p dns.DNSProvider, domain *email.Domain) (string, error) {
	// Build <0 issue "name"> list
	var records []dns.RRecord
	for _, ca := range domain.CAAs {
		record := dns.RRecord{
			Prio:  0,
			Value: fmt.Sprintf(`0 issue "%s"`, ca),
		}
		records = append(records, record)
	}

	msg := fmt.Sprintf("SetCAA %s -> %v (ttl=%d)", domain.Name, records, cfg.RegTTL)
	if cfg.Verbose {
		cfg.Debug(msg)
	} else if cfg.Dry {
		cfg.Say(msg)
		return "", nil
	}

	return p.UpsertRRSet(ctx, domain.Name, dns.RRSet{
		Name:    "@",
		Type:    dns.RR_CAA,
		TTL:     cfg.RegTTL,
		Records: records,
	})
}
