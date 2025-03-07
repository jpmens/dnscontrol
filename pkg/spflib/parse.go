package spflib

import (
	"errors"
	"fmt"
	"strings"
)

// SPFRecord stores the parts of an SPF record.
type SPFRecord struct {
	Parts []*SPFPart
}

// SPFPart stores a part of an SPF record, with attributes.
type SPFPart struct {
	Text          string
	IsLookup      bool
	IncludeRecord *SPFRecord
	IncludeDomain string
}

var qualifiers = map[byte]bool{
	'?': true,
	'~': true,
	'-': true,
	'+': true,
}

// Parse parses a raw SPF record.
func Parse(text string, dnsres Resolver) (*SPFRecord, error) {
	if !strings.HasPrefix(text, "v=spf1 ") {
		return nil, errors.New("not an SPF record")
	}
	parts := strings.Split(text, " ")
	rec := &SPFRecord{}
	for pi, part := range parts[1:] {
		if part == "" {
			continue
		}
		p := &SPFPart{Text: part}
		if qualifiers[part[0]] {
			part = part[1:]
		}
		rec.Parts = append(rec.Parts, p)
		if part == "all" {
			// all. nothing else matters.
			break
		} else if strings.HasPrefix(part, "a") || strings.HasPrefix(part, "mx") {
			p.IsLookup = true
		} else if strings.HasPrefix(part, "ip4:") || strings.HasPrefix(part, "ip6:") {
			// ip address, 0 lookups
			continue
		} else if strings.HasPrefix(part, "include:") || strings.HasPrefix(part, "redirect=") {
			// redirect is only partially implemented. redirect is a
			// complex and IMHO ambiguously defined feature.  We only
			// implement the most simple edge case: when it is the last item
			// in the string.  In that situation, it is the equivalent of
			// include:.
			if strings.HasPrefix(part, "redirect=") {
				// pi + 2: because pi starts at 0 when it iterates starting on parts[1],
				// and because len(parts) is one bigger than the highest index.
				if (pi + 2) != len(parts) {
					return nil, fmt.Errorf("%s must be last item", part)
				}
				p.IncludeDomain = strings.TrimPrefix(part, "redirect=")
			} else {
				p.IncludeDomain = strings.TrimPrefix(part, "include:")
			}
			p.IsLookup = true
			if dnsres != nil {
				subRecord, err := dnsres.GetSPF(p.IncludeDomain)
				if err != nil {
					return nil, err
				}
				p.IncludeRecord, err = Parse(subRecord, dnsres)
				if err != nil {
					return nil, fmt.Errorf("in included SPF: %w", err)
				}
			}
		} else if strings.HasPrefix(part, "exists:") || strings.HasPrefix(part, "ptr:") {
			p.IsLookup = true
		} else {
			return nil, fmt.Errorf("unsupported SPF part %s", part)
		}
	}
	return rec, nil
}
