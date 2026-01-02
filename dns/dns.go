package dns

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type RRSetType string

const (
	RR_A     RRSetType = "A"
	RR_AAAA  RRSetType = "AAAA"
	RR_CNAME RRSetType = "CNAME"
	RR_MX    RRSetType = "MX"
	RR_TXT   RRSetType = "TXT"
	RR_CAA   RRSetType = "CAA"
)

type RRecord struct {
	Prio  int
	Value string
	Text  string
}

type RRSet struct {
	Name    string // "@", "www", "_dmarc", "selector._domainkey"
	Type    RRSetType
	TTL     int
	Records []RRecord
}

type DNSProvider interface {
	UpsertRRSet(ctx context.Context, zone string, rr RRSet) (string, error)
}

type NoopProvider struct {
	Mode string
}

func NewNoopProvider(reason string) (*NoopProvider, error) {
	return &NoopProvider{
		Mode: reason,
	}, nil
}

func (p *NoopProvider) UpsertRRSet(ctx context.Context, zone string, rr RRSet) (string, error) {
	return fmt.Sprintf("⭕ %s (%s) %s for %s: %v", p.Mode, zone, rr.Type, rr.Name, rr.Records), nil
}

func normalizeZone(zone string) string {
	return strings.ToLower(strings.TrimSuffix(zone, "."))
}

func normalizeName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" || name == "@" {
		return "@"
	}

	// IMPORTANT: name must be relative to the zone, e.g. "www" not "www.example.com"
	return strings.ToLower(strings.TrimSuffix(name, "."))
}

func trimDot(s string) string {
	return strings.TrimSuffix(strings.TrimSpace(s), ".")
}

// normalizeRRSet prepares an RRSet for provider-specific processing.
//
// It must ensure that:
//   - rr.Records are sorted deterministically using the same ordering
//     as records returned by the provider.
//   - rr.Records[i].Text is the canonical comparison form used by checkSame().
func normalizeRRSet(rr RRSet) (RRSet, error) {
	rr.Name = normalizeName(rr.Name)

	out := make([]RRecord, 0, len(rr.Records))
	seen := make(map[string]struct{}, len(rr.Records))

	add := func(v RRecord, key string) {
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, v)
	}

	switch rr.Type {

	case RR_CNAME:
		for _, v := range rr.Records {
			v.Value = strings.TrimSpace(v.Value)
			if v.Value == "" {
				continue
			}
			v.Text = v.Value
			if !strings.HasSuffix(v.Text, ".") {
				v.Text += "."
			}

			add(v, v.Value)
		}
		if len(out) != 1 {
			return RRSet{}, fmt.Errorf("CNAME %q must have exactly one target, got %d", rr.Name, len(out))
		}

	case RR_MX:
		for _, v := range rr.Records {
			if v.Prio <= 0 {
				continue
			}
			v.Value = strings.TrimSpace(v.Value)
			if v.Value == "" {
				continue
			}
			v.Text = fmt.Sprintf("%d %s", v.Prio, v.Value)
			if !strings.HasSuffix(v.Text, ".") {
				v.Text += "."
			}

			add(v, v.Text) // key includes prio
		}
		if len(out) == 0 {
			return RRSet{}, fmt.Errorf("MX %q has no valid records", rr.Name)
		}
		sort.Slice(out, func(i, j int) bool {
			if out[i].Prio != out[j].Prio {
				return out[i].Prio < out[j].Prio
			}
			return out[i].Value < out[j].Value
		})

	case RR_TXT:
		for _, v := range rr.Records {
			v.Value = strings.TrimSpace(v.Value)
			if v.Value == "" {
				continue
			}
			v.Text = ValueToChunk(v.Value) // always: canonical 200-byte quoted chunks

			add(v, v.Value) // dedup by semantic value
		}
		sort.Slice(out, func(i, j int) bool {
			return out[i].Value < out[j].Value
		})

	default:
		for _, v := range rr.Records {
			v.Value = strings.TrimSpace(v.Value)
			if v.Value == "" {
				continue
			}
			v.Text = v.Value

			add(v, v.Value)
		}
		sort.Slice(out, func(i, j int) bool {
			return out[i].Value < out[j].Value
		})
	}

	rr.Records = out
	return rr, nil
}

// ValueToChunk converts a TXT value (unquoted, logical full string) into a
// provider-friendly chunked representation: `"part1" "part2" ...`
// Each part is at most 200 BYTES before quoting/escaping.
// This keeps you safely below the DNS 255-byte per character-string limit.
func ValueToChunk(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	const max = 200
	b := []byte(value)

	parts := make([]string, 0, (len(b)/max)+1)
	for len(b) > 0 {
		n := max
		if len(b) < n {
			n = len(b)
		}
		parts = append(parts, strconv.Quote(string(b[:n])))
		b = b[n:]
	}

	return strings.Join(parts, " ")
}

// ChunkToValue converts a chunked TXT representation (`"p1" "p2" ...`) back
// into the logical full string by unquoting each chunk and concatenating.
// If the input does not look like chunked quoted strings, it is returned trimmed.
// This is useful for canonical comparisons (no-op detection).
func ChunkToValue(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	// If it doesn't start with a quote, treat it as plain value.
	if !strings.HasPrefix(s, "\"") {
		return s
	}

	var out strings.Builder
	for {
		s = strings.TrimSpace(s)
		if s == "" {
			break
		}
		if s[0] != '"' {
			// Unexpected format; fall back.
			return strings.TrimSpace(s)
		}

		// Scan to the end quote, honoring backslash escapes.
		i := 1
		esc := false
		for i < len(s) {
			c := s[i]
			if esc {
				esc = false
				i++
				continue
			}
			if c == '\\' {
				esc = true
				i++
				continue
			}
			if c == '"' {
				break
			}
			i++
		}
		if i >= len(s) || s[i] != '"' {
			return strings.TrimSpace(s) // broken input
		}

		token := s[:i+1] // includes closing quote
		unq, err := strconv.Unquote(token)
		if err != nil {
			return strings.TrimSpace(s)
		}
		out.WriteString(unq)

		s = s[i+1:] // consume token
	}

	return out.String()
}

func canonTXTWire(s string) string {
	return ValueToChunk(ChunkToValue(strings.TrimSpace(s)))
}
