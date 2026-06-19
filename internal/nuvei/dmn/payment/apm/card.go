package apm

import "github.com/parkerscobey/nuvei-dmn-simulator/internal/nuvei/dmn/payment"

const (
	CardPaymentMethod = "cc_card"
	CardCurrency      = "USD"
	CardTotalAmount   = "30.00"
)

// Card builds a payment DMN payload using sanitized card payment defaults.
func Card(opts payment.Options) (payment.Payload, error) {
	if opts.PaymentMethod == "" {
		opts.PaymentMethod = CardPaymentMethod
	}
	if opts.TransactionType == "" {
		opts.TransactionType = "Sale"
	}
	if opts.Type == "" {
		opts.Type = "DEPOSIT"
	}
	if opts.Currency == "" {
		opts.Currency = CardCurrency
	}
	if opts.TotalAmount == "" {
		opts.TotalAmount = CardTotalAmount
	}
	if opts.NameOnCard == "" {
		opts.NameOnCard = "Test Cardholder"
	}
	if opts.CardNumber == "" {
		opts.CardNumber = "400002******0000"
	}
	if opts.ExpMonth == "" {
		opts.ExpMonth = "12"
	}
	if opts.ExpYear == "" {
		opts.ExpYear = "2030"
	}
	if opts.CardCompany == "" {
		opts.CardCompany = "Visa"
	}

	if opts.Status == payment.StatusDeclined {
		if opts.Reason == "" {
			opts.Reason = "Rejected by simulator."
		}
		if opts.ReasonCode == "" {
			opts.ReasonCode = "9999"
		}
		if opts.ErrCode == "" {
			opts.ErrCode = "9"
		}
	}

	return payment.Build(opts)
}
