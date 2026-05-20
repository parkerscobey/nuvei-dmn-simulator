package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestParseConfig(t *testing.T) {
	cfg, err := Parse([]byte(`
# Placeholder comment.
[merchants.local-demo]
environment = "test"
merchant_id = "merchant-id-placeholder"
merchant_site_id = "merchant-site-id-placeholder"
merchant_secret_key = "merchant-secret-key-placeholder"

[targets.local]
url = "http://localhost:3000/nuvei_direct_merchant_notifications"
kind = "local"
requires_confirm = false

[targets.prod_sandbox_gateway]
url = "https://app.example.com/nuvei_direct_merchant_notifications"
kind = "production-hosted-sandbox"
requires_confirm = true
`))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	merchant := cfg.Merchants["local-demo"]
	if merchant.Environment != "test" {
		t.Fatalf("Environment = %q, want test", merchant.Environment)
	}
	if merchant.MerchantID != "merchant-id-placeholder" {
		t.Fatalf("MerchantID = %q", merchant.MerchantID)
	}
	if merchant.MerchantSiteID != "merchant-site-id-placeholder" {
		t.Fatalf("MerchantSiteID = %q", merchant.MerchantSiteID)
	}
	if merchant.MerchantSecretKey != "merchant-secret-key-placeholder" {
		t.Fatalf("MerchantSecretKey = %q", merchant.MerchantSecretKey)
	}

	local := cfg.Targets["local"]
	if local.URL != "http://localhost:3000/nuvei_direct_merchant_notifications" {
		t.Fatalf("local URL = %q", local.URL)
	}
	if local.Kind != "local" {
		t.Fatalf("local Kind = %q", local.Kind)
	}
	if local.RequiresConfirm {
		t.Fatal("local RequiresConfirm = true, want false")
	}

	prodSandbox := cfg.Targets["prod_sandbox_gateway"]
	if !prodSandbox.RequiresConfirm {
		t.Fatal("prod_sandbox_gateway RequiresConfirm = false, want true")
	}
}

func TestRedactedDoesNotExposeSecrets(t *testing.T) {
	cfg := Empty()
	cfg.Merchants["local-demo"] = MerchantProfile{
		Environment:       "test",
		MerchantID:        "merchant-id-placeholder",
		MerchantSiteID:    "merchant-site-id-placeholder",
		MerchantSecretKey: "do-not-print-this",
	}

	formatted := Format(Redacted(cfg))
	if strings.Contains(formatted, "do-not-print-this") {
		t.Fatalf("redacted config exposed secret: %s", formatted)
	}
	if !strings.Contains(formatted, `merchant_secret_key = "`+RedactedSecret+`"`) {
		t.Fatalf("redacted config did not include redaction marker: %s", formatted)
	}
}

func TestSaveLoadConfig(t *testing.T) {
	cfg := Empty()
	cfg.Merchants["local-demo"] = MerchantProfile{
		Environment:       "test",
		MerchantID:        "merchant-id-placeholder",
		MerchantSiteID:    "merchant-site-id-placeholder",
		MerchantSecretKey: "merchant-secret-key-placeholder",
	}
	cfg.Targets["local"] = TargetProfile{
		URL:  "http://localhost:3000/nuvei_direct_merchant_notifications",
		Kind: "local",
	}

	path := filepath.Join(t.TempDir(), "nested", FileName)
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if loaded.Merchants["local-demo"].MerchantSecretKey != "merchant-secret-key-placeholder" {
		t.Fatalf("loaded merchant secret = %q", loaded.Merchants["local-demo"].MerchantSecretKey)
	}
	if loaded.Targets["local"].URL != "http://localhost:3000/nuvei_direct_merchant_notifications" {
		t.Fatalf("loaded target URL = %q", loaded.Targets["local"].URL)
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat returned error: %v", err)
		}
		if got := info.Mode().Perm(); got != 0o600 {
			t.Fatalf("config permissions = %o, want 600", got)
		}
	}
}

func TestParseRejectsUnknownKeys(t *testing.T) {
	_, err := Parse([]byte(`
[merchants.local-demo]
environment = "test"
merchant_secret_key = "merchant-secret-key-placeholder"
unexpected = "value"
`))
	if err == nil {
		t.Fatal("Parse returned nil error for unknown key")
	}
}

func TestValidateProfiles(t *testing.T) {
	if err := ValidateMerchantProfile(MerchantProfile{
		Environment:       "test",
		MerchantID:        "merchant-id-placeholder",
		MerchantSiteID:    "merchant-site-id-placeholder",
		MerchantSecretKey: "merchant-secret-key-placeholder",
	}); err != nil {
		t.Fatalf("ValidateMerchantProfile returned error: %v", err)
	}

	if err := ValidateMerchantProfile(MerchantProfile{Environment: "sandbox"}); err == nil {
		t.Fatal("ValidateMerchantProfile returned nil error for invalid environment")
	}

	if err := ValidateTargetProfile(TargetProfile{
		URL:  "http://localhost:3000/nuvei_direct_merchant_notifications",
		Kind: "local",
	}); err != nil {
		t.Fatalf("ValidateTargetProfile returned error: %v", err)
	}

	if err := ValidateTargetProfile(TargetProfile{URL: "localhost:3000", Kind: "local"}); err == nil {
		t.Fatal("ValidateTargetProfile returned nil error for relative URL")
	}
}
