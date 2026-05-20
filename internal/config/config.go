package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	DirName  = "nuvei-dmn-simulator"
	FileName = "config.toml"

	RedactedSecret = "********"
)

type Config struct {
	Merchants map[string]MerchantProfile
	Targets   map[string]TargetProfile
}

type MerchantProfile struct {
	Environment       string
	MerchantID        string
	MerchantSiteID    string
	MerchantSecretKey string
}

type TargetProfile struct {
	URL             string
	Kind            string
	RequiresConfirm bool
}

func Empty() Config {
	return Config{
		Merchants: map[string]MerchantProfile{},
		Targets:   map[string]TargetProfile{},
	}
}

func DefaultPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate user config directory: %w", err)
	}

	return filepath.Join(configDir, DirName, FileName), nil
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Empty(), nil
		}
		return Config{}, fmt.Errorf("read config %q: %w", path, err)
	}

	cfg, err := Parse(data)
	if err != nil {
		return Config{}, fmt.Errorf("parse config %q: %w", path, err)
	}

	return cfg, nil
}

func Save(path string, cfg Config) error {
	if err := validateProfileNames(cfg); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(Format(cfg)), 0o600); err != nil {
		return fmt.Errorf("write config %q: %w", path, err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return fmt.Errorf("restrict config permissions %q: %w", path, err)
	}

	return nil
}

func Parse(data []byte) (Config, error) {
	cfg := Empty()
	sectionKind := ""
	sectionName := ""

	for lineNumber, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(stripComment(rawLine))
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "[") {
			if !strings.HasSuffix(line, "]") {
				return Config{}, fmt.Errorf("line %d: invalid table header", lineNumber+1)
			}

			kind, name, err := parseSection(line[1 : len(line)-1])
			if err != nil {
				return Config{}, fmt.Errorf("line %d: %w", lineNumber+1, err)
			}
			sectionKind = kind
			sectionName = name

			if sectionKind == "merchants" {
				if _, ok := cfg.Merchants[sectionName]; !ok {
					cfg.Merchants[sectionName] = MerchantProfile{}
				}
			} else {
				if _, ok := cfg.Targets[sectionName]; !ok {
					cfg.Targets[sectionName] = TargetProfile{}
				}
			}
			continue
		}

		if sectionKind == "" {
			return Config{}, fmt.Errorf("line %d: key-value pair before a table header", lineNumber+1)
		}

		key, rawValue, ok := strings.Cut(line, "=")
		if !ok {
			return Config{}, fmt.Errorf("line %d: expected key = value", lineNumber+1)
		}
		key = strings.TrimSpace(key)
		value := strings.TrimSpace(rawValue)

		if sectionKind == "merchants" {
			profile := cfg.Merchants[sectionName]
			if err := applyMerchantValue(&profile, key, value); err != nil {
				return Config{}, fmt.Errorf("line %d: %w", lineNumber+1, err)
			}
			cfg.Merchants[sectionName] = profile
			continue
		}

		target := cfg.Targets[sectionName]
		if err := applyTargetValue(&target, key, value); err != nil {
			return Config{}, fmt.Errorf("line %d: %w", lineNumber+1, err)
		}
		cfg.Targets[sectionName] = target
	}

	return cfg, nil
}

func Format(cfg Config) string {
	var builder strings.Builder

	merchantNames := sortedKeys(cfg.Merchants)
	for _, name := range merchantNames {
		profile := cfg.Merchants[name]
		fmt.Fprintf(&builder, "[merchants.%s]\n", name)
		fmt.Fprintf(&builder, "environment = %s\n", quote(profile.Environment))
		fmt.Fprintf(&builder, "merchant_id = %s\n", quote(profile.MerchantID))
		fmt.Fprintf(&builder, "merchant_site_id = %s\n", quote(profile.MerchantSiteID))
		fmt.Fprintf(&builder, "merchant_secret_key = %s\n\n", quote(profile.MerchantSecretKey))
	}

	targetNames := sortedKeys(cfg.Targets)
	for _, name := range targetNames {
		target := cfg.Targets[name]
		fmt.Fprintf(&builder, "[targets.%s]\n", name)
		fmt.Fprintf(&builder, "url = %s\n", quote(target.URL))
		fmt.Fprintf(&builder, "kind = %s\n", quote(target.Kind))
		fmt.Fprintf(&builder, "requires_confirm = %t\n\n", target.RequiresConfirm)
	}

	return builder.String()
}

func Redacted(cfg Config) Config {
	redacted := Empty()
	for name, profile := range cfg.Merchants {
		if profile.MerchantSecretKey != "" {
			profile.MerchantSecretKey = RedactedSecret
		}
		redacted.Merchants[name] = profile
	}
	for name, target := range cfg.Targets {
		redacted.Targets[name] = target
	}

	return redacted
}

func ValidateMerchantProfile(profile MerchantProfile) error {
	switch profile.Environment {
	case "test", "prod":
	default:
		return fmt.Errorf("environment must be test or prod")
	}
	if profile.MerchantID == "" {
		return fmt.Errorf("merchant ID is required")
	}
	if profile.MerchantSiteID == "" {
		return fmt.Errorf("merchant site ID is required")
	}
	if profile.MerchantSecretKey == "" {
		return fmt.Errorf("merchant secret key is required")
	}

	return nil
}

func ValidateTargetProfile(target TargetProfile) error {
	if target.URL == "" {
		return fmt.Errorf("target URL is required")
	}
	parsed, err := url.Parse(target.URL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("target URL must be an absolute URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("target URL must use http or https")
	}
	if target.Kind == "" {
		return fmt.Errorf("target kind is required")
	}
	if !validTargetKind(target.Kind) {
		return fmt.Errorf("target kind must be local, staging, sandbox, demo, trusted, or production-hosted-sandbox")
	}

	return nil
}

func validTargetKind(kind string) bool {
	switch kind {
	case "local", "staging", "sandbox", "demo", "trusted", "production-hosted-sandbox":
		return true
	default:
		return false
	}
}

func parseSection(section string) (string, string, error) {
	section = strings.TrimSpace(section)
	if kind, name, ok := strings.Cut(section, "."); ok {
		if kind != "merchants" && kind != "targets" {
			return "", "", fmt.Errorf("unsupported table %q", section)
		}
		if !validProfileName(name) {
			return "", "", fmt.Errorf("invalid %s profile name %q", kind, name)
		}
		return kind, name, nil
	}

	return "", "", fmt.Errorf("unsupported table %q", section)
}

func applyMerchantValue(profile *MerchantProfile, key, rawValue string) error {
	value, err := parseStringValue(rawValue)
	if err != nil {
		return fmt.Errorf("%s must be a string: %w", key, err)
	}

	switch key {
	case "environment":
		profile.Environment = value
	case "merchant_id":
		profile.MerchantID = value
	case "merchant_site_id":
		profile.MerchantSiteID = value
	case "merchant_secret_key":
		profile.MerchantSecretKey = value
	default:
		return fmt.Errorf("unknown merchant key %q", key)
	}

	return nil
}

func applyTargetValue(target *TargetProfile, key, rawValue string) error {
	switch key {
	case "url":
		value, err := parseStringValue(rawValue)
		if err != nil {
			return fmt.Errorf("url must be a string: %w", err)
		}
		target.URL = value
	case "kind":
		value, err := parseStringValue(rawValue)
		if err != nil {
			return fmt.Errorf("kind must be a string: %w", err)
		}
		target.Kind = value
	case "requires_confirm":
		value, err := parseBoolValue(rawValue)
		if err != nil {
			return fmt.Errorf("requires_confirm must be a boolean: %w", err)
		}
		target.RequiresConfirm = value
	default:
		return fmt.Errorf("unknown target key %q", key)
	}

	return nil
}

func parseStringValue(rawValue string) (string, error) {
	if !strings.HasPrefix(rawValue, "\"") {
		return "", fmt.Errorf("expected quoted string")
	}
	value, err := strconv.Unquote(rawValue)
	if err != nil {
		return "", err
	}

	return value, nil
}

func parseBoolValue(rawValue string) (bool, error) {
	switch rawValue {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("expected true or false")
	}
}

func stripComment(line string) string {
	inString := false
	escaped := false
	for i, char := range line {
		if escaped {
			escaped = false
			continue
		}
		if inString && char == '\\' {
			escaped = true
			continue
		}
		if char == '"' {
			inString = !inString
			continue
		}
		if !inString && char == '#' {
			return line[:i]
		}
	}

	return line
}

func validateProfileNames(cfg Config) error {
	for name := range cfg.Merchants {
		if !validProfileName(name) {
			return fmt.Errorf("invalid merchant profile name %q", name)
		}
	}
	for name := range cfg.Targets {
		if !validProfileName(name) {
			return fmt.Errorf("invalid target profile name %q", name)
		}
	}

	return nil
}

func validProfileName(name string) bool {
	if name == "" {
		return false
	}
	for _, char := range name {
		if char >= 'a' && char <= 'z' {
			continue
		}
		if char >= 'A' && char <= 'Z' {
			continue
		}
		if char >= '0' && char <= '9' {
			continue
		}
		if char == '-' || char == '_' {
			continue
		}
		return false
	}

	return true
}

func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func quote(value string) string {
	return strconv.Quote(value)
}
