package targetsafe

import (
	"errors"
	"strings"
	"testing"
)

type mockConfirmer struct {
	confirm      bool
	confirmError error
	called       bool
	lastPrompt   string
}

func (m *mockConfirmer) Confirm(prompt string) (bool, error) {
	m.called = true
	m.lastPrompt = prompt
	return m.confirm, m.confirmError
}

func TestRequireAllowed_LocalAllowed(t *testing.T) {
	t.Parallel()
	r := Result{Classification: ClassificationLocal, URL: "http://localhost:3000"}
	err := RequireAllowed(r, false, nil)
	if err != nil {
		t.Errorf("RequireAllowed(local, false, nil) = %v, want nil", err)
	}
}

func TestRequireAllowed_StagingAllowed(t *testing.T) {
	t.Parallel()
	r := Result{Classification: ClassificationStaging, URL: "https://app.staging/nuvei"}
	err := RequireAllowed(r, false, nil)
	if err != nil {
		t.Errorf("RequireAllowed(staging, false, nil) = %v, want nil", err)
	}
}

func TestRequireAllowed_UntrustedNeedsFlag(t *testing.T) {
	t.Parallel()
	r := Result{Classification: ClassificationUntrusted, URL: "https://example.com/nuvei", Reason: "unknown public host"}

	err := RequireAllowed(r, false, nil)
	if err == nil {
		t.Fatal("RequireAllowed(untrusted, false, nil) = nil, want error")
	}
	if !strings.Contains(err.Error(), "not allowed by default") {
		t.Errorf("error = %v, want 'not allowed by default' message", err)
	}

	err = RequireAllowed(r, true, nil)
	if err != nil {
		t.Errorf("RequireAllowed(untrusted, true, nil) = %v, want nil", err)
	}
}

func TestRequireAllowed_DeniedAlwaysDenied(t *testing.T) {
	t.Parallel()
	r := Result{Classification: ClassificationDenied, URL: "file:///etc/passwd", Reason: "denied URL scheme"}

	err := RequireAllowed(r, true, nil)
	if err == nil {
		t.Error("RequireAllowed(denied, true, nil) = nil, want error")
	}
	if !strings.Contains(err.Error(), "is denied") {
		t.Errorf("error = %v, want 'is denied' message", err)
	}
}

func TestRequireAllowed_ProdSandboxNeedsConfirm(t *testing.T) {
	t.Parallel()
	r := Result{Classification: ClassificationProductionHostedSandbox, URL: "https://app.example.com/nuvei", Reason: "trusted profile prod-sandbox", RequiresConfirm: true}

	confirmer := &mockConfirmer{confirm: true}
	err := RequireAllowed(r, false, confirmer)
	if err != nil {
		t.Errorf("RequireAllowed(prod-sandbox, false, confirm-yes) = %v, want nil", err)
	}
	if !confirmer.called {
		t.Error("confirmer was not called")
	}
}

func TestRequireAllowed_ConfirmCancelled(t *testing.T) {
	t.Parallel()
	r := Result{Classification: ClassificationProductionHostedSandbox, URL: "https://app.example.com/nuvei", Reason: "trusted profile prod-sandbox", RequiresConfirm: true}

	confirmer := &mockConfirmer{confirm: false}
	err := RequireAllowed(r, false, confirmer)
	if err == nil {
		t.Error("RequireAllowed(prod-sandbox, false, confirm-no) = nil, want error")
	}
	if !strings.Contains(err.Error(), "cancelled") {
		t.Errorf("error = %v, want 'cancelled'", err)
	}
}

func TestRequireAllowed_ConfirmError(t *testing.T) {
	t.Parallel()
	r := Result{Classification: ClassificationProductionHostedSandbox, URL: "https://app.example.com/nuvei", Reason: "trusted profile", RequiresConfirm: true}

	confirmer := &mockConfirmer{confirmError: errors.New("stdin closed")}
	err := RequireAllowed(r, false, confirmer)
	if err == nil {
		t.Error("RequireAllowed(confirm-error) = nil, want error")
	}
	if !strings.Contains(err.Error(), "confirm prompt") {
		t.Errorf("error = %v, want 'confirm prompt' message", err)
	}
}

func TestRequireAllowed_UntrustedWithSuggestion(t *testing.T) {
	t.Parallel()
	r := Result{Classification: ClassificationUntrusted, URL: "https://unknown.com/nuvei", Reason: "unknown public HTTPS host"}
	err := RequireAllowed(r, false, nil)
	if err == nil {
		t.Fatal("got nil error")
	}
	if !strings.Contains(err.Error(), "--allow-untrusted-target") {
		t.Errorf("error = %v, want suggestion about --allow-untrusted-target", err)
	}
}

func TestRequireAllowed_AllowUntrustedSilencesUntrusted(t *testing.T) {
	t.Parallel()
	r := Result{Classification: ClassificationUntrusted, URL: "https://example.com/nuvei", Reason: "unknown"}
	err := RequireAllowed(r, true, nil)
	if err != nil {
		t.Errorf("RequireAllowed(untrusted, true, nil) = %v, want nil", err)
	}
}

func TestRequireAllowed_ProdSandboxNoConfirmer(t *testing.T) {
	t.Parallel()
	r := Result{Classification: ClassificationProductionHostedSandbox, URL: "https://app.example.com/nuvei", Reason: "trusted profile", RequiresConfirm: true}

	err := RequireAllowed(r, false, nil)
	if err == nil {
		t.Error("RequireAllowed(prod-sandbox, false, nil) = nil, want error")
	}
	if !strings.Contains(err.Error(), "requires confirmation") {
		t.Errorf("error = %v, want 'requires confirmation'", err)
	}
}

func TestCheck_LocalAllowed(t *testing.T) {
	t.Parallel()
	result, err := Check("http://localhost:3000/nuvei", CheckOptions{})
	if err != nil {
		t.Fatalf("Check(local) error = %v, want nil", err)
	}
	if result.Classification != ClassificationLocal {
		t.Fatalf("Check(local) classification = %v, want local", result.Classification)
	}
}

func TestCheck_UntrustedBlockedByDefault(t *testing.T) {
	t.Parallel()
	_, err := Check("https://example.com/nuvei", CheckOptions{})
	if err == nil {
		t.Fatal("Check(untrusted) error = nil, want error")
	}
}

func TestCheck_AllowUntrustedAllowsUnknownPublicHost(t *testing.T) {
	t.Parallel()
	result, err := Check("https://example.com/nuvei", CheckOptions{AllowUntrusted: true})
	if err != nil {
		t.Fatalf("Check(untrusted, allow) error = %v, want nil", err)
	}
	if result.Classification != ClassificationUntrusted {
		t.Fatalf("Check(untrusted, allow) classification = %v, want untrusted", result.Classification)
	}
}

func TestCheck_DeniedNotBypassedByAllowUntrusted(t *testing.T) {
	t.Parallel()
	_, err := Check("file:///etc/passwd", CheckOptions{AllowUntrusted: true})
	if err == nil {
		t.Fatal("Check(denied, allow) error = nil, want error")
	}
}

func TestCheck_ProductionHostedSandboxRequiresConfirmation(t *testing.T) {
	t.Parallel()
	profiles := map[string]Profile{
		"prod-sandbox": {URL: "https://app.example.com/nuvei", Kind: "production-hosted-sandbox"},
	}
	confirmer := &mockConfirmer{confirm: true}
	result, err := Check("https://app.example.com/nuvei", CheckOptions{TrustedProfiles: profiles, Confirmer: confirmer})
	if err != nil {
		t.Fatalf("Check(prod-sandbox) error = %v, want nil", err)
	}
	if result.Classification != ClassificationProductionHostedSandbox {
		t.Fatalf("Check(prod-sandbox) classification = %v, want production-hosted-sandbox", result.Classification)
	}
	if !confirmer.called {
		t.Fatal("Check(prod-sandbox) did not prompt for confirmation")
	}
}
