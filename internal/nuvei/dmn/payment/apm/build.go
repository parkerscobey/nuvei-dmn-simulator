package apm

import "github.com/parkerscobey/nuvei-dmn-simulator/internal/nuvei/dmn/payment"

// Defaults defines reusable APM-level defaults for payment DMN payloads.
type Defaults struct {
	PaymentMethod      string
	TransactionType    string
	Type               string
	Currency           string
	TotalAmount        string
	DeclinedReason     string
	DeclinedReasonCode string
	DeclinedErrCode    string
}

// Build applies APM defaults and then builds a signed payment DMN payload.
func Build(opts payment.Options, defaults Defaults) (payment.Payload, error) {
	if opts.PaymentMethod == "" {
		opts.PaymentMethod = defaults.PaymentMethod
	}
	if opts.TransactionType == "" {
		opts.TransactionType = defaults.TransactionType
	}
	if opts.Type == "" {
		opts.Type = defaults.Type
	}
	if opts.Currency == "" {
		opts.Currency = defaults.Currency
	}
	if opts.TotalAmount == "" {
		opts.TotalAmount = defaults.TotalAmount
	}

	if opts.Status == payment.StatusDeclined {
		if opts.Reason == "" {
			opts.Reason = defaults.DeclinedReason
		}
		if opts.ReasonCode == "" {
			opts.ReasonCode = defaults.DeclinedReasonCode
		}
		if opts.ErrCode == "" {
			opts.ErrCode = defaults.DeclinedErrCode
		}
	}

	return payment.Build(opts)
}
