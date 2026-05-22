package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/parkerscobey/nuvei-dmn-simulator/internal/nuvei/credentials"
	"github.com/parkerscobey/nuvei-dmn-simulator/internal/sender"
)

func TestUIIntegrationVerifyPreviewSendFlow(t *testing.T) {
	t.Parallel()

	configPath := writeConfig(t)
	verifyCalls := 0
	sendCalls := 0
	var lastPayload string

	h, err := NewHandler(
		configPath,
		func(context.Context, credentials.Profile) (credentials.Verification, error) {
			verifyCalls++
			return credentials.Verification{Environment: "test"}, nil
		},
		func(_ context.Context, _ string, encodedPayload string) (sender.Result, error) {
			sendCalls++
			lastPayload = encodedPayload
			return sender.Result{StatusCode: http.StatusCreated, Body: "accepted by receiver"}, nil
		},
	)
	if err != nil {
		t.Fatalf("NewHandler error = %v", err)
	}

	server := httptest.NewServer(h.Routes())
	defer server.Close()

	httpClient := server.Client()

	homeResp, err := httpClient.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("GET / error = %v", err)
	}
	homeBody := mustReadBody(t, homeResp)
	if homeResp.StatusCode != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", homeResp.StatusCode)
	}
	if !strings.Contains(homeBody, "Nuvei DMN Simulator") {
		t.Fatalf("GET / missing page title: %s", homeBody)
	}

	baseForm := url.Values{}
	baseForm.Set("profile", "local-demo")
	baseForm.Set("target", "local")
	baseForm.Set("apm", "pix")
	baseForm.Set("status", "APPROVED")

	verifyResp, err := httpClient.PostForm(server.URL+"/htmx/verify", baseForm)
	if err != nil {
		t.Fatalf("POST /htmx/verify error = %v", err)
	}
	verifyBody := mustReadBody(t, verifyResp)
	if verifyResp.StatusCode != http.StatusOK {
		t.Fatalf("verify status = %d, want 200", verifyResp.StatusCode)
	}
	if !strings.Contains(verifyBody, "Verified profile") {
		t.Fatalf("verify response missing success message: %s", verifyBody)
	}

	previewResp, err := httpClient.PostForm(server.URL+"/htmx/preview", baseForm)
	if err != nil {
		t.Fatalf("POST /htmx/preview error = %v", err)
	}
	previewBody := mustReadBody(t, previewResp)
	if previewResp.StatusCode != http.StatusOK {
		t.Fatalf("preview status = %d, want 200", previewResp.StatusCode)
	}
	if !strings.Contains(previewBody, "Raw URL-encoded payload") {
		t.Fatalf("preview response missing payload section: %s", previewBody)
	}

	sendResp, err := httpClient.PostForm(server.URL+"/htmx/send", baseForm)
	if err != nil {
		t.Fatalf("POST /htmx/send error = %v", err)
	}
	sendBody := mustReadBody(t, sendResp)
	if sendResp.StatusCode != http.StatusOK {
		t.Fatalf("send status = %d, want 200", sendResp.StatusCode)
	}
	if !strings.Contains(sendBody, "Status: 201") {
		t.Fatalf("send response missing status: %s", sendBody)
	}
	if !strings.Contains(sendBody, "accepted by receiver") {
		t.Fatalf("send response missing body: %s", sendBody)
	}

	if verifyCalls != 2 {
		t.Fatalf("verify calls = %d, want 2 (verify + send)", verifyCalls)
	}
	if sendCalls != 1 {
		t.Fatalf("send calls = %d, want 1", sendCalls)
	}
	if !strings.Contains(lastPayload, "advanceResponseChecksum=") {
		t.Fatalf("last payload missing checksum: %s", lastPayload)
	}
}

func mustReadBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll body error = %v", err)
	}
	return string(body)
}
