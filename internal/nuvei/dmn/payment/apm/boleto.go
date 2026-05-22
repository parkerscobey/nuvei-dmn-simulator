package apm

import "github.com/parkerscobey/nuvei-dmn-simulator/internal/nuvei/dmn/payment"

const (
	BoletoPaymentMethod = "apmgw_BOLETO"
	BoletoCurrency      = "BRL"
	BoletoTotalAmount   = "30.00"
)

func boletoDefaults() Defaults {
	return Defaults{
		PaymentMethod:      BoletoPaymentMethod,
		TransactionType:    "Sale",
		Type:               "DEPOSIT",
		Currency:           BoletoCurrency,
		TotalAmount:        BoletoTotalAmount,
		DeclinedReason:     "Rejected by simulator.",
		DeclinedReasonCode: "9999",
		DeclinedErrCode:    "9",
	}
}

// Boleto builds a payment DMN payload using Boleto defaults.
func Boleto(opts payment.Options) (payment.Payload, error) {
	return Build(opts, boletoDefaults())
}
