package payment

import (
	"encoding/hex"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestPPPStatusForStatus(t *testing.T) {
	tests := map[string]string{
		StatusPending:  PPPStatusPending,
		StatusApproved: PPPStatusOK,
		StatusDeclined: PPPStatusFail,
	}

	for status, want := range tests {
		got, err := PPPStatusForStatus(status)
		if err != nil {
			t.Fatalf("PPPStatusForStatus(%q) error = %v", status, err)
		}
		if got != want {
			t.Fatalf("PPPStatusForStatus(%q) = %q, want %q", status, got, want)
		}
	}
}

func TestPPPStatusForStatusRejectsUnknownStatus(t *testing.T) {
	if _, err := PPPStatusForStatus("SUCCESS"); err == nil {
		t.Fatal("PPPStatusForStatus(SUCCESS) error = nil, want error")
	}
}

func TestBuildPaymentPayload(t *testing.T) {
	payload, err := Build(Options{
		MerchantID:          "merchant-id",
		MerchantSiteID:      "site-id",
		MerchantSecretKey:   "secret",
		TotalAmount:         "30.00",
		Currency:            "BRL",
		ResponseTimeStamp:   "2026-05-20.18:10:00",
		PPPTransactionID:    "1234567890",
		Status:              StatusApproved,
		ProductID:           "PIX-DEMO",
		PaymentMethod:       "apmgw_PIX",
		TransactionType:     "Sale",
		Type:                "DEPOSIT",
		ClientRequestID:     "req-1",
		ClientUniqueID:      "uniq-1",
		UserPaymentOptionID: "upo-1",
	})
	if err != nil {
		t.Fatalf("Build error = %v", err)
	}

	assertField(t, payload, FieldPPPStatus, PPPStatusOK)
	assertField(t, payload, FieldMessage, StatusApproved)
	assertField(t, payload, FieldAdvanceResponseChecksum, "d85119067f127635b8a8b4720bbc780330fff7607bfeb2defc7a73881990c2df")

	encoded := payload.Encode()
	values, err := url.ParseQuery(encoded)
	if err != nil {
		t.Fatalf("ParseQuery(%q) error = %v", encoded, err)
	}
	if values.Get(FieldMerchantID) != "merchant-id" {
		t.Fatalf("encoded merchant_id = %q, want merchant-id", values.Get(FieldMerchantID))
	}
	if values.Get(FieldAdvanceResponseChecksum) == "" {
		t.Fatal("encoded advanceResponseChecksum is empty")
	}
}

func TestBuildGeneratesMissingIDsAndTimestamp(t *testing.T) {
	payload, err := Build(Options{
		MerchantID:        "merchant-id",
		MerchantSiteID:    "site-id",
		MerchantSecretKey: "secret",
		TotalAmount:       "30.00",
		Currency:          "BRL",
		Status:            StatusPending,
		ProductID:         "",
		PaymentMethod:     "apmgw_PIX",
		TransactionType:   "Sale",
		Type:              "DEPOSIT",
		Now: func() time.Time {
			return time.Date(2026, 5, 20, 18, 10, 0, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatalf("Build error = %v", err)
	}

	assertField(t, payload, FieldResponseTimeStamp, "2026-05-20.18:10:00")
	expected := map[string]string{
		FieldPPPTransactionID:    "ppp-20260520181000",
		FieldClientRequestID:     "req-20260520181000",
		FieldClientUniqueID:      "uniq-20260520181000",
		FieldUserPaymentOptionID: "upo-20260520181000",
	}
	for key, wantPrefix := range expected {
		assertGeneratedID(t, payload, key, wantPrefix)
	}
}

func TestBuildRejectsMissingRequiredField(t *testing.T) {
	_, err := Build(Options{
		MerchantSiteID:    "site-id",
		MerchantSecretKey: "secret",
		TotalAmount:       "30.00",
		Currency:          "BRL",
		Status:            StatusApproved,
		ProductID:         "",
		PaymentMethod:     "apmgw_PIX",
		TransactionType:   "Sale",
		Type:              "DEPOSIT",
	})
	if err == nil {
		t.Fatal("Build error = nil, want missing merchant_id error")
	}
}

func assertField(t *testing.T, payload Payload, key, want string) {
	t.Helper()
	if got := payload.Fields[key]; got != want {
		t.Fatalf("field %q = %q, want %q", key, got, want)
	}
}

func assertGeneratedID(t *testing.T, payload Payload, key, wantPrefix string) {
	t.Helper()
	got := payload.Fields[key]
	prefix := wantPrefix + "-"
	if !strings.HasPrefix(got, prefix) {
		t.Fatalf("field %q = %q, want prefix %q", key, got, prefix)
	}
	suffix := strings.TrimPrefix(got, prefix)
	if len(suffix) != 8 {
		t.Fatalf("field %q suffix = %q, want 8 hex characters", key, suffix)
	}
	if _, err := hex.DecodeString(suffix); err != nil {
		t.Fatalf("field %q suffix = %q, want hex: %v", key, suffix, err)
	}
}

func TestParseEncodedParsesAndValidatesPayload(t *testing.T) {
	t.Parallel()

	raw := "merchant_id=merchant-id&merchant_site_id=site-id&totalAmount=30.00&currency=BRL&responseTimeStamp=2026-05-20.18%3A10%3A00&PPP_TransactionID=ppp-1&Status=APPROVED&productId=&payment_method=apmgw_PIX&ppp_status=OK&message=APPROVED&transactionType=Sale&type=DEPOSIT&clientRequestId=req-1&clientUniqueId=uniq-1&userPaymentOptionId=upo-1&advanceResponseChecksum=checksum"

	payload, err := ParseEncoded(raw)
	if err != nil {
		t.Fatalf("ParseEncoded error = %v", err)
	}

	assertField(t, payload, FieldMerchantID, "merchant-id")
	assertField(t, payload, FieldStatus, StatusApproved)
}

func TestRecomputeAdvanceResponseChecksumUpdatesChecksum(t *testing.T) {
	t.Parallel()

	payload := Payload{Fields: map[string]string{
		FieldMerchantID:              "merchant-id",
		FieldMerchantSiteID:          "site-id",
		FieldTotalAmount:             "30.00",
		FieldCurrency:                "BRL",
		FieldResponseTimeStamp:       "2026-05-20.18:10:00",
		FieldPPPTransactionID:        "1234567890",
		FieldStatus:                  StatusApproved,
		FieldProductID:               "PIX-DEMO",
		FieldPaymentMethod:           "apmgw_PIX",
		FieldPPPStatus:               PPPStatusOK,
		FieldMessage:                 StatusApproved,
		FieldTransactionType:         "Sale",
		FieldType:                    "DEPOSIT",
		FieldClientRequestID:         "req-1",
		FieldClientUniqueID:          "uniq-1",
		FieldUserPaymentOptionID:     "upo-1",
		FieldAdvanceResponseChecksum: "old",
	}}

	recomputed, err := RecomputeAdvanceResponseChecksum(payload, "secret")
	if err != nil {
		t.Fatalf("RecomputeAdvanceResponseChecksum error = %v", err)
	}

	if recomputed.Fields[FieldAdvanceResponseChecksum] == "old" {
		t.Fatal("checksum was not recomputed")
	}
	if !strings.EqualFold(recomputed.Fields[FieldAdvanceResponseChecksum], "d85119067f127635b8a8b4720bbc780330fff7607bfeb2defc7a73881990c2df") {
		t.Fatalf("unexpected checksum = %q", recomputed.Fields[FieldAdvanceResponseChecksum])
	}
}
