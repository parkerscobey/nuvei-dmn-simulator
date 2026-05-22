package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	appconfig "github.com/parkerscobey/nuvei-dmn-simulator/internal/config"
	"github.com/parkerscobey/nuvei-dmn-simulator/internal/nuvei/credentials"
	"github.com/parkerscobey/nuvei-dmn-simulator/internal/sender"
)

func TestHomeRendersForm(t *testing.T) {
	t.Parallel()

	h := mustHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.Routes().ServeHTTP(w, r)

	body := w.Body.String()
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(body, "Merchant profile") {
		t.Fatalf("missing profile label: %s", body)
	}
	if !strings.Contains(body, "hx-post=\"/htmx/preview\"") {
		t.Fatalf("missing preview htmx action: %s", body)
	}
	if !strings.Contains(body, "<option value=\"boleto\">boleto</option>") {
		t.Fatalf("missing boleto APM option: %s", body)
	}
}

func TestPreviewEndpointRendersPayload(t *testing.T) {
	t.Parallel()

	h := mustHandler(t)
	form := url.Values{}
	form.Set("profile", "local-demo")
	form.Set("target", "local")
	form.Set("status", "APPROVED")

	r := httptest.NewRequest(http.MethodPost, "/htmx/preview", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.Routes().ServeHTTP(w, r)

	body := w.Body.String()
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(body, "Raw URL-encoded payload") {
		t.Fatalf("missing raw payload heading: %s", body)
	}
	if !strings.Contains(body, "advanceResponseChecksum") {
		t.Fatalf("missing checksum: %s", body)
	}
}

func TestPreviewEndpointRendersBoletoPayload(t *testing.T) {
	t.Parallel()

	h := mustHandler(t)
	form := url.Values{}
	form.Set("profile", "local-demo")
	form.Set("target", "local")
	form.Set("status", "PENDING")
	form.Set("apm", "boleto")

	r := httptest.NewRequest(http.MethodPost, "/htmx/preview", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.Routes().ServeHTTP(w, r)

	body := w.Body.String()
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(body, "apmgw_BOLETO") {
		t.Fatalf("missing boleto payload marker: %s", body)
	}
}

func TestSendEndpointUsesVerifyAndSender(t *testing.T) {
	t.Parallel()

	configPath := writeConfig(t)
	var verified bool
	var sent bool

	h, err := NewHandler(
		configPath,
		func(context.Context, credentials.Profile) (credentials.Verification, error) {
			verified = true
			return credentials.Verification{Environment: "test"}, nil
		},
		func(context.Context, string, string) (sender.Result, error) {
			sent = true
			return sender.Result{StatusCode: http.StatusAccepted, Body: "ok"}, nil
		},
	)
	if err != nil {
		t.Fatalf("NewHandler error = %v", err)
	}

	form := url.Values{}
	form.Set("profile", "local-demo")
	form.Set("target", "local")
	form.Set("status", "APPROVED")

	r := httptest.NewRequest(http.MethodPost, "/htmx/send", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.Routes().ServeHTTP(w, r)

	body := w.Body.String()
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !verified {
		t.Fatal("verify function was not called")
	}
	if !sent {
		t.Fatal("send function was not called")
	}
	if !strings.Contains(body, "Status: 202") {
		t.Fatalf("missing HTTP status: %s", body)
	}
}

func mustHandler(t *testing.T) *Handler {
	t.Helper()
	h, err := NewHandler(
		writeConfig(t),
		func(context.Context, credentials.Profile) (credentials.Verification, error) {
			return credentials.Verification{Environment: "test"}, nil
		},
		func(context.Context, string, string) (sender.Result, error) {
			return sender.Result{StatusCode: http.StatusOK, Body: "ok"}, nil
		},
	)
	if err != nil {
		t.Fatalf("NewHandler error = %v", err)
	}
	return h
}

func writeConfig(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	err := appconfig.Save(path, appconfig.Config{
		Merchants: map[string]appconfig.MerchantProfile{
			"local-demo": {
				Environment:       "test",
				MerchantID:        "merchant-id-placeholder",
				MerchantSiteID:    "merchant-site-id-placeholder",
				MerchantSecretKey: "merchant-secret-key-placeholder",
			},
		},
		Targets: map[string]appconfig.TargetProfile{
			"local": {
				URL:  "http://localhost:3000/nuvei_direct_merchant_notifications",
				Kind: "local",
			},
		},
	})
	if err != nil {
		t.Fatalf("Save config error = %v", err)
	}
	return path
}
