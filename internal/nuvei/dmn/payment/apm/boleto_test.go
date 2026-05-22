package apm

import (
	"testing"

	"github.com/parkerscobey/nuvei-dmn-simulator/internal/nuvei/dmn/payment"
)

func TestBoletoDefaults(t *testing.T) {
	payload, err := Boleto(payment.Options{
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
		t.Fatalf("Boleto error = %v", err)
	}

	assertField(t, payload, payment.FieldPaymentMethod, BoletoPaymentMethod)
	assertField(t, payload, payment.FieldTransactionType, "Sale")
	assertField(t, payload, payment.FieldType, "DEPOSIT")
	assertField(t, payload, payment.FieldCurrency, BoletoCurrency)
	assertField(t, payload, payment.FieldTotalAmount, BoletoTotalAmount)
	assertField(t, payload, payment.FieldProductID, "")
	assertField(t, payload, payment.FieldMessage, payment.StatusApproved)
	assertField(t, payload, payment.FieldPPPStatus, payment.PPPStatusOK)
}

func TestBoletoStatusMappings(t *testing.T) {
	tests := map[string]string{
		payment.StatusPending:  payment.PPPStatusPending,
		payment.StatusApproved: payment.PPPStatusOK,
		payment.StatusDeclined: payment.PPPStatusFail,
	}

	for status, wantPPPStatus := range tests {
		payload, err := Boleto(payment.Options{
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
			t.Fatalf("Boleto(%q) error = %v", status, err)
		}
		assertField(t, payload, payment.FieldPPPStatus, wantPPPStatus)
		assertField(t, payload, payment.FieldMessage, status)
	}
}

func TestBoletoDeclinedDefaults(t *testing.T) {
	payload, err := Boleto(payment.Options{
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
		t.Fatalf("Boleto error = %v", err)
	}

	assertField(t, payload, payment.FieldReason, "Rejected by simulator.")
	assertField(t, payload, payment.FieldReasonCode, "9999")
	assertField(t, payload, payment.FieldErrCode, "9")
}
