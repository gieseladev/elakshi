package infoextract

import (
	"net/url"
	"strings"
)

// HasHostnames is an interface for an Extractor which extracts for specific
// hostnames.
type HasHostnames interface {
	Extractor
	// URLHostnames returns a slice of hostnames.
	// The hostnames shouldn't include "www.".
	URLHostnames() []string
}

// A URLChecker is an Extractor which extracts from URLs.
type URLChecker interface {
	Extractor

	// CheckURL checks whether the extractor works for the given url.
	CheckURL(u *url.URL) bool
}

// A URIChecker is an Extractor which extracts URIs.
type URIChecker interface {
	Extractor

	// CheckURI checks whether the extractor works for the given uri.
	CheckURI(uri string) bool
}

// An ExtractorPool collects extractors.
type ExtractorPool struct {
	extractors map[string]Extractor

	uriCheckers         []URIChecker
	urlCheckers         []URLChecker
	extractorByHostname map[string]Extractor
}

// Create a new extractor pool from the given extractors.
// Passing a nil extractor will cause a panic.
func CollectExtractors(extractors ...Extractor) *ExtractorPool {
	p := new(ExtractorPool)

	p.extractors = make(map[string]Extractor)
	p.extractorByHostname = make(map[string]Extractor)

	for _, extractor := range extractors {
		if extractor == nil {
			panic("nil extractor passed")
		}

		extractorID := extractor.ExtractorID()
		p.extractors[extractorID] = extractor

		if checker, ok := extractor.(URIChecker); ok {
			p.uriCheckers = append(p.uriCheckers, checker)
		}

		if checker, ok := extractor.(HasHostnames); ok {
			for _, hostname := range checker.URLHostnames() {
				p.extractorByHostname[hostname] = checker
			}
		}

		if checker, ok := extractor.(URLChecker); ok {
			p.urlCheckers = append(p.urlCheckers, checker)
		}
	}

	return p
}

func (p *ExtractorPool) resolveExtractorFromURL(u *url.URL) (Extractor, bool) {
	hostname := strings.TrimPrefix(u.Hostname(), "www.")
	extractor, ok := p.extractorByHostname[hostname]
	if ok {
		return extractor, true
	}

	for _, checker := range p.urlCheckers {
		if checker.CheckURL(u) {
			return checker, true
		}
	}

	return nil, false
}

// ResolveExtractor finds an extractor for the given uri.
// To be discovered, an extractor has to register using AddCheckers.
func (p *ExtractorPool) ResolveExtractor(uri string) (Extractor, bool) {
	if u, err := url.Parse(uri); err == nil && u.Host != "" {
		return p.resolveExtractorFromURL(u)
	}

	for _, checker := range p.uriCheckers {
		if checker.CheckURI(uri) {
			return checker, true
		}
	}

	return nil, false
}

// GetExtractor returns the extractor with the given extractor id.
func (p *ExtractorPool) GetExtractor(extractorID string) (Extractor, bool) {
	e, ok := p.extractors[extractorID]
	return e, ok
}
