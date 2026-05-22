package apm

import (
	"testing"

	"github.com/parkerscobey/nuvei-dmn-simulator/internal/nuvei/dmn/payment"
)

func TestBuildAppliesGenericAPMDefaults(t *testing.T) {
	payload, err := Build(payment.Options{
		MerchantID:          "merchant-id",
		MerchantSiteID:      "site-id",
		MerchantSecretKey:   "secret",
		ResponseTimeStamp:   "2026-05-20.18:10:00",
		PPPTransactionID:    "1234567890",
		Status:              payment.StatusDeclined,
		ClientRequestID:     "req-1",
		ClientUniqueID:      "uniq-1",
		UserPaymentOptionID: "upo-1",
	}, Defaults{
		PaymentMethod:      "apmgw_GENERIC",
		TransactionType:    "Sale",
		Type:               "DEPOSIT",
		Currency:           "BRL",
		TotalAmount:        "45.00",
		DeclinedReason:     "declined",
		DeclinedReasonCode: "1000",
		DeclinedErrCode:    "10",
	})
	if err != nil {
		t.Fatalf("Build error = %v", err)
	}

	assertField(t, payload, payment.FieldPaymentMethod, "apmgw_GENERIC")
	assertField(t, payload, payment.FieldCurrency, "BRL")
	assertField(t, payload, payment.FieldTotalAmount, "45.00")
	assertField(t, payload, payment.FieldReason, "declined")
	assertField(t, payload, payment.FieldReasonCode, "1000")
	assertField(t, payload, payment.FieldErrCode, "10")
}
