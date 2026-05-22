package apm

import "github.com/parkerscobey/nuvei-dmn-simulator/internal/nuvei/dmn/payment"

const (
	PixPaymentMethod = "apmgw_PIX"
	PixCurrency      = "BRL"
	PixTotalAmount   = "30.00"
)

func pixDefaults() Defaults {
	return Defaults{
		PaymentMethod:      PixPaymentMethod,
		TransactionType:    "Sale",
		Type:               "DEPOSIT",
		Currency:           PixCurrency,
		TotalAmount:        PixTotalAmount,
		DeclinedReason:     "Rejected by simulator.",
		DeclinedReasonCode: "9999",
		DeclinedErrCode:    "9",
	}
}

// Pix builds a payment DMN payload using Pix defaults.
func Pix(opts payment.Options) (payment.Payload, error) {
	return Build(opts, pixDefaults())
}
