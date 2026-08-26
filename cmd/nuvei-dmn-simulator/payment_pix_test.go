package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	appconfig "github.com/parkerscobey/nuvei-dmn-simulator/internal/config"
	"github.com/parkerscobey/nuvei-dmn-simulator/internal/nuvei/credentials"
	"github.com/parkerscobey/nuvei-dmn-simulator/internal/sender"
)

func TestPreviewPaymentPixPrintsTableAndRawPayload(t *testing.T) {
	t.Parallel()

	configPath := writeTestConfig(t, appconfig.Config{
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

	cmd := newRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{
		"preview", "payment", "pix",
		"--config", configPath,
		"--profile", "local-demo",
		"--target", "local",
		"--status", "APPROVED",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Payload fields:") {
		t.Fatalf("output missing payload table header: %s", output)
	}
	if !strings.Contains(output, "Raw URL-encoded payload:") {
		t.Fatalf("output missing raw payload header: %s", output)
	}
	if !strings.Contains(output, "advanceResponseChecksum") {
		t.Fatalf("output missing checksum field: %s", output)
	}
	if !strings.Contains(output, "merchant_id=merchant-id-placeholder") {
		t.Fatalf("output missing raw URL-encoded payload body: %s", output)
	}
}

func TestPreviewPaymentBoletoPrintsTableAndRawPayload(t *testing.T) {
	t.Parallel()

	configPath := writeTestConfig(t, appconfig.Config{
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

	cmd := newRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{
		"preview", "payment", "boleto",
		"--config", configPath,
		"--profile", "local-demo",
		"--target", "local",
		"--status", "PENDING",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Payload fields:") {
		t.Fatalf("output missing payload table header: %s", output)
	}
	if !strings.Contains(output, "Raw URL-encoded payload:") {
		t.Fatalf("output missing raw payload header: %s", output)
	}
	if !strings.Contains(output, "payment_method           apmgw_BOLETO") {
		t.Fatalf("output missing boleto payment method: %s", output)
	}
}

func TestPreviewPaymentCardPrintsTableAndRawPayload(t *testing.T) {
	t.Parallel()

	configPath := writeTestConfig(t, appconfig.Config{
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

	cmd := newRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{
		"preview", "payment", "card",
		"--config", configPath,
		"--profile", "local-demo",
		"--target", "local",
		"--status", "APPROVED",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "payment_method           cc_card") {
		t.Fatalf("output missing card payment method: %s", output)
	}
	if !strings.Contains(output, "cardNumber") || !strings.Contains(output, "400002******0000") {
		t.Fatalf("output missing masked card number: %s", output)
	}
}

func TestPreviewPaymentLocalPaymentsAfricaPrintsTableAndRawPayload(t *testing.T) {
	t.Parallel()

	configPath := writeTestConfig(t, appconfig.Config{
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

	cmd := newRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{
		"preview", "payment", "local-payments-africa",
		"--config", configPath,
		"--profile", "local-demo",
		"--target", "local",
		"--status", "PENDING",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Payload fields:") {
		t.Fatalf("output missing payload table header: %s", output)
	}
	if !strings.Contains(output, "Raw URL-encoded payload:") {
		t.Fatalf("output missing raw payload header: %s", output)
	}
	if !strings.Contains(output, "payment_method           apmgw_Local_payments_Africa") {
		t.Fatalf("output missing local payments Africa payment method: %s", output)
	}
}

func TestSendPaymentPixPostsToTargetAndPrintsStatusAndBody(t *testing.T) {
	var (
		requestMethod      string
		requestContentType string
		requestBody        string
	)

	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestMethod = r.Method
		requestContentType = r.Header.Get("Content-Type")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll request body error = %v", err)
		}
		requestBody = string(body)

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("accepted by target"))
	}))
	defer targetServer.Close()

	configPath := writeTestConfig(t, appconfig.Config{
		Merchants: map[string]appconfig.MerchantProfile{
			"local-demo": {
				Environment:       "test",
				MerchantID:        "merchant-id-placeholder",
				MerchantSiteID:    "merchant-site-id-placeholder",
				MerchantSecretKey: "merchant-secret-key-placeholder",
			},
		},
		Targets: map[string]appconfig.TargetProfile{},
	})

	originalVerify := verifyMerchantProfile
	verifyMerchantProfile = func(ctx context.Context, profile credentials.Profile) (credentials.Verification, error) {
		return credentials.Verification{Environment: profile.Environment}, nil
	}
	t.Cleanup(func() {
		verifyMerchantProfile = originalVerify
	})

	cmd := newRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{
		"send", "payment", "pix",
		"--config", configPath,
		"--profile", "local-demo",
		"--target", targetServer.URL,
		"--status", "APPROVED",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute error = %v", err)
	}

	if requestMethod != http.MethodPost {
		t.Fatalf("method = %q, want %q", requestMethod, http.MethodPost)
	}
	if requestContentType != sender.ContentTypeFormURLEncoded {
		t.Fatalf("Content-Type = %q, want %q", requestContentType, sender.ContentTypeFormURLEncoded)
	}
	if !strings.Contains(requestBody, "advanceResponseChecksum=") {
		t.Fatalf("request body missing checksum: %s", requestBody)
	}

	output := stdout.String()
	if !strings.Contains(output, "HTTP status: 201") {
		t.Fatalf("output missing status code: %s", output)
	}
	if !strings.Contains(output, "accepted by target") {
		t.Fatalf("output missing response body: %s", output)
	}
}

func TestSendPaymentBoletoPostsToTargetAndPrintsStatusAndBody(t *testing.T) {
	var requestBody string
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll request body error = %v", err)
		}
		requestBody = string(body)

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("accepted by target"))
	}))
	defer targetServer.Close()

	configPath := writeTestConfig(t, appconfig.Config{
		Merchants: map[string]appconfig.MerchantProfile{
			"local-demo": {
				Environment:       "test",
				MerchantID:        "merchant-id-placeholder",
				MerchantSiteID:    "merchant-site-id-placeholder",
				MerchantSecretKey: "merchant-secret-key-placeholder",
			},
		},
		Targets: map[string]appconfig.TargetProfile{},
	})

	originalVerify := verifyMerchantProfile
	verifyMerchantProfile = func(ctx context.Context, profile credentials.Profile) (credentials.Verification, error) {
		return credentials.Verification{Environment: profile.Environment}, nil
	}
	t.Cleanup(func() {
		verifyMerchantProfile = originalVerify
	})

	cmd := newRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{
		"send", "payment", "boleto",
		"--config", configPath,
		"--profile", "local-demo",
		"--target", targetServer.URL,
		"--status", "APPROVED",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute error = %v", err)
	}

	if !strings.Contains(requestBody, "payment_method=apmgw_BOLETO") {
		t.Fatalf("request body missing boleto payment method: %s", requestBody)
	}
	output := stdout.String()
	if !strings.Contains(output, "HTTP status: 201") {
		t.Fatalf("output missing status code: %s", output)
	}
	if !strings.Contains(output, "accepted by target") {
		t.Fatalf("output missing response body: %s", output)
	}
}

func TestSendPaymentCardPostsToTargetAndPrintsStatusAndBody(t *testing.T) {
	var requestBody string
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll request body error = %v", err)
		}
		requestBody = string(body)

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("accepted by target"))
	}))
	defer targetServer.Close()

	configPath := writeTestConfig(t, appconfig.Config{
		Merchants: map[string]appconfig.MerchantProfile{
			"local-demo": {
				Environment:       "test",
				MerchantID:        "merchant-id-placeholder",
				MerchantSiteID:    "merchant-site-id-placeholder",
				MerchantSecretKey: "merchant-secret-key-placeholder",
			},
		},
		Targets: map[string]appconfig.TargetProfile{},
	})

	originalVerify := verifyMerchantProfile
	verifyMerchantProfile = func(ctx context.Context, profile credentials.Profile) (credentials.Verification, error) {
		return credentials.Verification{Environment: profile.Environment}, nil
	}
	t.Cleanup(func() {
		verifyMerchantProfile = originalVerify
	})

	cmd := newRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{
		"send", "payment", "card",
		"--config", configPath,
		"--profile", "local-demo",
		"--target", targetServer.URL,
		"--status", "APPROVED",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute error = %v", err)
	}

	if !strings.Contains(requestBody, "payment_method=cc_card") {
		t.Fatalf("request body missing card payment method: %s", requestBody)
	}
	if !strings.Contains(requestBody, "cardNumber=400002%2A%2A%2A%2A%2A%2A0000") {
		t.Fatalf("request body missing masked card number: %s", requestBody)
	}
	output := stdout.String()
	if !strings.Contains(output, "HTTP status: 201") {
		t.Fatalf("output missing status code: %s", output)
	}
	if !strings.Contains(output, "accepted by target") {
		t.Fatalf("output missing response body: %s", output)
	}
}

func TestSendPaymentLocalPaymentsAfricaPostsToTargetAndPrintsStatusAndBody(t *testing.T) {
	var requestBody string
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll request body error = %v", err)
		}
		requestBody = string(body)

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("accepted by target"))
	}))
	defer targetServer.Close()

	configPath := writeTestConfig(t, appconfig.Config{
		Merchants: map[string]appconfig.MerchantProfile{
			"local-demo": {
				Environment:       "test",
				MerchantID:        "merchant-id-placeholder",
				MerchantSiteID:    "merchant-site-id-placeholder",
				MerchantSecretKey: "merchant-secret-key-placeholder",
			},
		},
		Targets: map[string]appconfig.TargetProfile{},
	})

	originalVerify := verifyMerchantProfile
	verifyMerchantProfile = func(ctx context.Context, profile credentials.Profile) (credentials.Verification, error) {
		return credentials.Verification{Environment: profile.Environment}, nil
	}
	t.Cleanup(func() {
		verifyMerchantProfile = originalVerify
	})

	cmd := newRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{
		"send", "payment", "local-payments-africa",
		"--config", configPath,
		"--profile", "local-demo",
		"--target", targetServer.URL,
		"--status", "APPROVED",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute error = %v", err)
	}

	if !strings.Contains(requestBody, "payment_method=apmgw_Local_payments_Africa") {
		t.Fatalf("request body missing local payments Africa payment method: %s", requestBody)
	}
	output := stdout.String()
	if !strings.Contains(output, "HTTP status: 201") {
		t.Fatalf("output missing status code: %s", output)
	}
	if !strings.Contains(output, "accepted by target") {
		t.Fatalf("output missing response body: %s", output)
	}
}

func TestSendPaymentPixBlocksUntrustedTargetByDefault(t *testing.T) {
	t.Parallel()

	configPath := writeTestConfig(t, appconfig.Config{
		Merchants: map[string]appconfig.MerchantProfile{
			"local-demo": {
				Environment:       "test",
				MerchantID:        "merchant-id-placeholder",
				MerchantSiteID:    "merchant-site-id-placeholder",
				MerchantSecretKey: "merchant-secret-key-placeholder",
			},
		},
		Targets: map[string]appconfig.TargetProfile{},
	})

	cmd := newRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{
		"send", "payment", "pix",
		"--config", configPath,
		"--profile", "local-demo",
		"--target", "https://example.com/nuvei_direct_merchant_notifications",
		"--status", "APPROVED",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute error = nil, want safety error")
	}
	if !strings.Contains(err.Error(), "not allowed") {
		t.Fatalf("error = %v, want untrusted target error", err)
	}
}

func TestPreviewPaymentPixStrictModeRequiresCorrelationFields(t *testing.T) {
	t.Parallel()

	configPath := writeTestConfig(t, appconfig.Config{
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

	cmd := newRootCommand()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{
		"preview", "payment", "pix",
		"--config", configPath,
		"--profile", "local-demo",
		"--target", "local",
		"--require-correlation-fields",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute error = nil, want strict mode validation error")
	}
	if !strings.Contains(err.Error(), "--total-amount") {
		t.Fatalf("error = %v, want missing --total-amount", err)
	}
}

func TestPreviewPaymentPixStrictModePassesWithRequiredFields(t *testing.T) {
	t.Parallel()

	configPath := writeTestConfig(t, appconfig.Config{
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

	cmd := newRootCommand()
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{
		"preview", "payment", "pix",
		"--config", configPath,
		"--profile", "local-demo",
		"--target", "local",
		"--require-correlation-fields",
		"--status", "APPROVED",
		"--total-amount", "42.10",
		"--currency", "BRL",
		"--user-payment-option-id", "upo-123",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute error = %v", err)
	}
}

func TestPreviewPaymentFromRawReplacesMerchantAndRecomputesChecksum(t *testing.T) {
	t.Parallel()

	configPath := writeTestConfig(t, appconfig.Config{
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

	rawFile := writeRawPayloadFile(t, "merchant_id=nuvei-docs-merchant&merchant_site_id=nuvei-docs-site&totalAmount=30.00&currency=BRL&responseTimeStamp=2026-05-20.18%3A10%3A00&PPP_TransactionID=1234567890&Status=APPROVED&productId=PIX-DEMO&payment_method=apmgw_PIX&ppp_status=OK&message=APPROVED&transactionType=Sale&type=DEPOSIT&clientRequestId=req-1&clientUniqueId=uniq-1&userPaymentOptionId=upo-1&advanceResponseChecksum=old")

	cmd := newRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{
		"preview", "payment", "from-raw",
		"--config", configPath,
		"--profile", "local-demo",
		"--file", rawFile,
		"--target", "local",
		"--status", "DECLINED",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute error = %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "merchant_id=merchant-id-placeholder") {
		t.Fatalf("output missing replaced merchant_id: %s", output)
	}
	if !strings.Contains(output, "Status                   DECLINED") {
		t.Fatalf("output missing status override: %s", output)
	}
	if strings.Contains(output, "advanceResponseChecksum\told") {
		t.Fatalf("checksum was not recomputed: %s", output)
	}
}

func TestSendPaymentFromRawPostsPayload(t *testing.T) {
	var requestBody string
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll request body error = %v", err)
		}
		requestBody = string(body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer targetServer.Close()

	configPath := writeTestConfig(t, appconfig.Config{
		Merchants: map[string]appconfig.MerchantProfile{
			"local-demo": {
				Environment:       "test",
				MerchantID:        "merchant-id-placeholder",
				MerchantSiteID:    "merchant-site-id-placeholder",
				MerchantSecretKey: "merchant-secret-key-placeholder",
			},
		},
		Targets: map[string]appconfig.TargetProfile{},
	})

	rawFile := writeRawPayloadFile(t, "merchant_id=nuvei-docs-merchant&merchant_site_id=nuvei-docs-site&totalAmount=30.00&currency=BRL&responseTimeStamp=2026-05-20.18%3A10%3A00&PPP_TransactionID=1234567890&Status=APPROVED&productId=PIX-DEMO&payment_method=apmgw_PIX&ppp_status=OK&message=APPROVED&transactionType=Sale&type=DEPOSIT&clientRequestId=req-1&clientUniqueId=uniq-1&userPaymentOptionId=upo-1&advanceResponseChecksum=old")

	originalVerify := verifyMerchantProfile
	verifyMerchantProfile = func(ctx context.Context, profile credentials.Profile) (credentials.Verification, error) {
		return credentials.Verification{Environment: profile.Environment}, nil
	}
	t.Cleanup(func() {
		verifyMerchantProfile = originalVerify
	})

	cmd := newRootCommand()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs([]string{
		"send", "payment", "from-raw",
		"--config", configPath,
		"--profile", "local-demo",
		"--file", rawFile,
		"--target", targetServer.URL,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute error = %v", err)
	}

	if !strings.Contains(requestBody, "merchant_id=merchant-id-placeholder") {
		t.Fatalf("request body missing replaced merchant_id: %s", requestBody)
	}
	if strings.Contains(requestBody, "advanceResponseChecksum=old") {
		t.Fatalf("request body has stale checksum: %s", requestBody)
	}
}

func writeTestConfig(t *testing.T, cfg appconfig.Config) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.toml")
	if err := appconfig.Save(path, cfg); err != nil {
		t.Fatalf("Save config error = %v", err)
	}
	return path
}

func writeRawPayloadFile(t *testing.T, rawPayload string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "payload.txt")
	if err := os.WriteFile(path, []byte(rawPayload), 0o600); err != nil {
		t.Fatalf("WriteFile payload error = %v", err)
	}
	return path
}
