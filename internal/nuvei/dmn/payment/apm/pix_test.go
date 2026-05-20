package apm

import (
	"testing"

	"github.com/parkerscobey/nuvei-dmn-simulator/internal/nuvei/dmn/payment"
)

func TestPixDefaults(t *testing.T) {
	payload, err := Pix(payment.Options{
		MerchantID:          "merchant-id",
		MerchantSiteID:      "site-id",
		MerchantSecretKey:   "secret",
		ResponseTimeStamp:   "2026-05-20.18:10:00",
		PPPTransactionID:    "1234567890",
		Status:              payment.StatusApproved,
		ClientRequestID:     "req-1",
		ClientUniqueID:      "uniq-1",
		UserPaymentOptionID: "upo-1",
	})
	if err != nil {
		t.Fatalf("Pix error = %v", err)
	}

	assertField(t, payload, payment.FieldPaymentMethod, PixPaymentMethod)
	assertField(t, payload, payment.FieldTransactionType, "Sale")
	assertField(t, payload, payment.FieldType, "DEPOSIT")
	assertField(t, payload, payment.FieldCurrency, PixCurrency)
	assertField(t, payload, payment.FieldTotalAmount, PixTotalAmount)
	assertField(t, payload, payment.FieldProductID, "")
	assertField(t, payload, payment.FieldMessage, payment.StatusApproved)
	assertField(t, payload, payment.FieldPPPStatus, payment.PPPStatusOK)
}

func TestPixStatusMappings(t *testing.T) {
	tests := map[string]string{
		payment.StatusPending:  payment.PPPStatusPending,
		payment.StatusApproved: payment.PPPStatusOK,
		payment.StatusDeclined: payment.PPPStatusFail,
	}

	for status, wantPPPStatus := range tests {
		payload, err := Pix(payment.Options{
			MerchantID:          "merchant-id",
			MerchantSiteID:      "site-id",
			MerchantSecretKey:   "secret",
			ResponseTimeStamp:   "2026-05-20.18:10:00",
			PPPTransactionID:    "1234567890",
			Status:              status,
			ClientRequestID:     "req-1",
			ClientUniqueID:      "uniq-1",
			UserPaymentOptionID: "upo-1",
		})
		if err != nil {
			t.Fatalf("Pix(%q) error = %v", status, err)
		}
		assertField(t, payload, payment.FieldPPPStatus, wantPPPStatus)
		assertField(t, payload, payment.FieldMessage, status)
	}
}

func TestPixDeclinedDefaults(t *testing.T) {
	payload, err := Pix(payment.Options{
		MerchantID:          "merchant-id",
		MerchantSiteID:      "site-id",
		MerchantSecretKey:   "secret",
		ResponseTimeStamp:   "2026-05-20.18:10:00",
		PPPTransactionID:    "1234567890",
		Status:              payment.StatusDeclined,
		ClientRequestID:     "req-1",
		ClientUniqueID:      "uniq-1",
		UserPaymentOptionID: "upo-1",
	})
	if err != nil {
		t.Fatalf("Pix error = %v", err)
	}

	assertField(t, payload, payment.FieldReason, "Rejected by simulator.")
	assertField(t, payload, payment.FieldReasonCode, "9999")
	assertField(t, payload, payment.FieldErrCode, "9")
}

func assertField(t *testing.T, payload payment.Payload, key, want string) {
	t.Helper()
	if got := payload.Fields[key]; got != want {
		t.Fatalf("field %q = %q, want %q", key, got, want)
	}
}
