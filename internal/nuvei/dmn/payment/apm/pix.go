package apm

import "github.com/parkerscobey/nuvei-dmn-simulator/internal/nuvei/dmn/payment"

const (
	PixPaymentMethod = "apmgw_PIX"
	PixCurrency      = "BRL"
	PixTotalAmount   = "30.00"
)

// Pix builds a payment DMN payload using Pix defaults.
func Pix(opts payment.Options) (payment.Payload, error) {
	if opts.PaymentMethod == "" {
		opts.PaymentMethod = PixPaymentMethod
	}
	if opts.TransactionType == "" {
		opts.TransactionType = "Sale"
	}
	if opts.Type == "" {
		opts.Type = "DEPOSIT"
	}
	if opts.Currency == "" {
		opts.Currency = PixCurrency
	}
	if opts.TotalAmount == "" {
		opts.TotalAmount = PixTotalAmount
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
