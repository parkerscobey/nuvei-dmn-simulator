package checksum

import "testing"

func TestPaymentSourceString(t *testing.T) {
	fields := PaymentFields{
		TotalAmount:       "30.00",
		Currency:          "BRL",
		ResponseTimeStamp: "2026-05-20.18:10:00",
		PPPTransactionID:  "1234567890",
		Status:            "APPROVED",
		ProductID:         "",
		MerchantSecretKey: "secret",
	}

	got := PaymentSourceString(fields)
	want := "30.00BRL2026-05-20.18:10:001234567890APPROVEDsecret"
	if got != want {
		t.Fatalf("source string = %q, want %q", got, want)
	}
}

func TestPaymentAdvanceResponseChecksum(t *testing.T) {
	fields := PaymentFields{
		TotalAmount:       "30.00",
		Currency:          "BRL",
		ResponseTimeStamp: "2026-05-20.18:10:00",
		PPPTransactionID:  "1234567890",
		Status:            "APPROVED",
		ProductID:         "PIX-DEMO",
		MerchantSecretKey: "secret",
	}

	got := PaymentAdvanceResponseChecksum(fields)
	want := "d85119067f127635b8a8b4720bbc780330fff7607bfeb2defc7a73881990c2df"
	if got != want {
		t.Fatalf("checksum = %q, want %q", got, want)
	}
}
