package payment

import (
	"net/url"
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
	for key, want := range expected {
		assertField(t, payload, key, want)
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
