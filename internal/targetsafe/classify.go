package targetsafe

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

type Classification uint8

const (
	ClassificationUnknown Classification = iota
	ClassificationLocal
	ClassificationStaging
	ClassificationProductionHostedSandbox
	ClassificationUntrusted
	ClassificationDenied
)

func (c Classification) String() string {
	switch c {
	case ClassificationLocal:
		return "local"
	case ClassificationStaging:
		return "staging"
	case ClassificationProductionHostedSandbox:
		return "production-hosted-sandbox"
	case ClassificationUntrusted:
		return "untrusted"
	case ClassificationDenied:
		return "denied"
	default:
		return "unknown"
	}
}

type Result struct {
	Classification  Classification
	Host            string
	URL             string
	Reason          string
	RequiresConfirm bool
}

type Profile struct {
	URL             string
	Kind            string
	RequiresConfirm bool
}

func (r Result) Allowed() bool {
	return r.Classification != ClassificationUntrusted &&
		r.Classification != ClassificationDenied &&
		r.Classification != ClassificationUnknown
}

var defaultDeniedPrefixes = []string{
	"file://",
	"javascript:",
	"data:",
}

var localSuffixes = []string{
	".local",
	".test",
	".localhost",
}

func Classify(targetURL string, trustedProfiles map[string]Profile, deniedPrefixes []string) Result {
	parsed, err := url.Parse(targetURL)
	if err != nil {
		return Result{
			Classification: ClassificationDenied,
			URL:            targetURL,
			Reason:         fmt.Sprintf("unparseable URL: %v", err),
		}
	}

	host := normalizeHost(parsed.Hostname())
	if host == "" {
		return Result{
			Classification: ClassificationDenied,
			URL:            targetURL,
			Reason:         "target URL must include a host",
		}
	}

	for _, prefix := range deniedPrefixes {
		if strings.HasPrefix(targetURL, prefix) {
			return Result{
				Classification: ClassificationDenied,
				Host:           host,
				URL:            targetURL,
				Reason:         "denied URL scheme",
			}
		}
	}
	for _, prefix := range defaultDeniedPrefixes {
		if strings.HasPrefix(targetURL, prefix) {
			return Result{
				Classification: ClassificationDenied,
				Host:           host,
				URL:            targetURL,
				Reason:         "denied URL scheme",
			}
		}
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return Result{
			Classification: ClassificationDenied,
			Host:           host,
			URL:            targetURL,
			Reason:         "only http/https are allowed",
		}
	}

	if host == "localhost" {
		return Result{
			Classification: ClassificationLocal,
			Host:           host,
			URL:            targetURL,
			Reason:         "localhost address",
		}
	}

	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		return Result{
			Classification: ClassificationLocal,
			Host:           host,
			URL:            targetURL,
			Reason:         "loopback address",
		}
	}

	for _, suffix := range localSuffixes {
		if strings.HasSuffix(host, suffix) {
			return Result{
				Classification: ClassificationLocal,
				Host:           host,
				URL:            targetURL,
				Reason:         "local TLD suffix",
			}
		}
	}

	if trustedProfiles != nil {
		for name, profile := range trustedProfiles {
			trusted, err := url.Parse(profile.URL)
			if err != nil {
				continue
			}
			trustedHost := normalizeHost(trusted.Hostname())
			if trustedHost == host {
				if profile.Kind == "production-hosted-sandbox" || profile.RequiresConfirm {
					return Result{
						Classification:  ClassificationProductionHostedSandbox,
						Host:            host,
						URL:             targetURL,
						Reason:          fmt.Sprintf("trusted profile %q (production-hosted-sandbox)", name),
						RequiresConfirm: true,
					}
				}
				return Result{
					Classification: classificationForProfileKind(profile.Kind),
					Host:           host,
					URL:            targetURL,
					Reason:         fmt.Sprintf("trusted profile %q", name),
				}
			}
		}
	}

	if scheme == "http" {
		return Result{
			Classification: ClassificationUntrusted,
			Host:           host,
			URL:            targetURL,
			Reason:         "non-loopback HTTP URL without trusted profile",
		}
	}

	return Result{
		Classification: ClassificationUntrusted,
		Host:           host,
		URL:            targetURL,
		Reason:         "unknown public HTTPS host without trusted profile",
	}
}

func normalizeHost(host string) string {
	return strings.TrimSuffix(strings.ToLower(host), ".")
}

func classificationForProfileKind(kind string) Classification {
	switch kind {
	case "local":
		return ClassificationLocal
	case "production-hosted-sandbox":
		return ClassificationProductionHostedSandbox
	case "staging", "sandbox", "demo", "trusted":
		return ClassificationStaging
	default:
		return ClassificationStaging
	}
}
