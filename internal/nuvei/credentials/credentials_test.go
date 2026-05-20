package credentials

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestVerifyUsesGetSessionTokenAndCachesSuccess(t *testing.T) {
	profile := Profile{
		Environment:       EnvironmentTest,
		MerchantID:        "merchant-id-placeholder",
		MerchantSiteID:    "merchant-site-id-placeholder",
		MerchantSecretKey: "merchant-secret-key-placeholder",
	}

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/ppp/api/v1/getSessionToken.do" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("Content-Type = %q", r.Header.Get("Content-Type"))
		}

		var request getSessionTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if request.MerchantID != profile.MerchantID {
			t.Fatalf("merchantId = %q", request.MerchantID)
		}
		if request.MerchantSiteID != profile.MerchantSiteID {
			t.Fatalf("merchantSiteId = %q", request.MerchantSiteID)
		}
		if request.ClientRequestID != "client-request-id-placeholder" {
			t.Fatalf("clientRequestId = %q", request.ClientRequestID)
		}
		if request.TimeStamp != "20260520123456" {
			t.Fatalf("timeStamp = %q", request.TimeStamp)
		}
		wantChecksum := getSessionTokenChecksum(profile, request.ClientRequestID, request.TimeStamp)
		if request.Checksum != wantChecksum {
			t.Fatalf("checksum = %q, want %q", request.Checksum, wantChecksum)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
"sessionToken":"session-token-placeholder",
"internalRequestId":123,
"status":"SUCCESS",
"errCode":0,
"reason":"",
"merchantId":"merchant-id-placeholder",
"merchantSiteId":"merchant-site-id-placeholder",
"version":"1.0",
"clientRequestId":"client-request-id-placeholder"
}`))
	}))
	defer server.Close()

	client := &Client{
		HTTPClient: server.Client(),
		Clock: func() time.Time {
			return time.Date(2026, 5, 20, 12, 34, 56, 0, time.UTC)
		},
		NewID:     func() (string, error) { return "client-request-id-placeholder", nil },
		Endpoints: Endpoints{Test: server.URL + "/ppp/api/v1/getSessionToken.do"},
		Cache:     NewSessionCache(),
	}

	verification, err := client.Verify(context.Background(), profile)
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
	if verification.Method != MethodGetSessionToken {
		t.Fatalf("Method = %q", verification.Method)
	}
	if verification.SessionToken != "session-token-placeholder" {
		t.Fatalf("SessionToken = %q", verification.SessionToken)
	}
	if verification.Cached {
		t.Fatal("first verification was cached")
	}
	if err := client.RequireVerified(profile); err != nil {
		t.Fatalf("RequireVerified returned error after verify: %v", err)
	}

	verification, err = client.Verify(context.Background(), profile)
	if err != nil {
		t.Fatalf("cached Verify returned error: %v", err)
	}
	if !verification.Cached {
		t.Fatal("second verification was not cached")
	}
	if requests != 1 {
		t.Fatalf("requests = %d, want 1", requests)
	}
}

func TestVerifyReturnsNuveiFailureReason(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ERROR","errCode":101,"reason":"failure in checksum validation"}`))
	}))
	defer server.Close()

	client := &Client{
		HTTPClient: server.Client(),
		Clock:      func() time.Time { return time.Date(2026, 5, 20, 12, 34, 56, 0, time.UTC) },
		NewID:      func() (string, error) { return "client-request-id-placeholder", nil },
		Endpoints:  Endpoints{Test: server.URL},
		Cache:      NewSessionCache(),
	}

	_, err := client.Verify(context.Background(), Profile{
		Environment:       EnvironmentTest,
		MerchantID:        "merchant-id-placeholder",
		MerchantSiteID:    "merchant-site-id-placeholder",
		MerchantSecretKey: "merchant-secret-key-placeholder",
	})
	if err == nil {
		t.Fatal("Verify returned nil error")
	}
	if got := err.Error(); got != "Nuvei credential verification failed: status=ERROR errCode=101 reason=failure in checksum validation" {
		t.Fatalf("error = %q", got)
	}
}

func TestRequireVerifiedRejectsUnverifiedProfile(t *testing.T) {
	client := &Client{Cache: NewSessionCache()}
	err := client.RequireVerified(Profile{
		Environment:       EnvironmentTest,
		MerchantID:        "merchant-id-placeholder",
		MerchantSiteID:    "merchant-site-id-placeholder",
		MerchantSecretKey: "merchant-secret-key-placeholder",
	})
	if err != ErrNotVerified {
		t.Fatalf("RequireVerified error = %v, want ErrNotVerified", err)
	}
}

func TestValidateProfileRejectsInvalidProfile(t *testing.T) {
	if err := ValidateProfile(Profile{Environment: "sandbox"}); err == nil {
		t.Fatal("ValidateProfile returned nil error for invalid environment")
	}
}
