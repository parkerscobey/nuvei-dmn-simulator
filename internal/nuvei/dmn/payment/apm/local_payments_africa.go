package apm

import "github.com/parkerscobey/nuvei-dmn-simulator/internal/nuvei/dmn/payment"

const (
	LocalPaymentsAfricaPaymentMethod = "apmgw_Local_payments_Africa"
	LocalPaymentsAfricaCurrency      = "USD"
	LocalPaymentsAfricaTotalAmount   = "30.00"
)

func localPaymentsAfricaDefaults() Defaults {
	return Defaults{
		PaymentMethod:      LocalPaymentsAfricaPaymentMethod,
		TransactionType:    "Sale",
		Type:               "DEPOSIT",
		Currency:           LocalPaymentsAfricaCurrency,
		TotalAmount:        LocalPaymentsAfricaTotalAmount,
		DeclinedReason:     "Rejected by simulator.",
		DeclinedReasonCode: "9999",
		DeclinedErrCode:    "9",
	}
}

// LocalPaymentsAfrica builds a payment DMN payload using Local Payments Africa defaults.
func LocalPaymentsAfrica(opts payment.Options) (payment.Payload, error) {
	return Build(opts, localPaymentsAfricaDefaults())
}
