package payment

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/parkerscobey/nuvei-dmn-simulator/internal/nuvei/checksum"
)

const (
	StatusPending  = "PENDING"
	StatusApproved = "APPROVED"
	StatusDeclined = "DECLINED"

	PPPStatusPending = "PENDING"
	PPPStatusOK      = "OK"
	PPPStatusFail    = "FAIL"

	FieldMerchantID              = "merchant_id"
	FieldMerchantSiteID          = "merchant_site_id"
	FieldTotalAmount             = "totalAmount"
	FieldCurrency                = "currency"
	FieldResponseTimeStamp       = "responseTimeStamp"
	FieldPPPTransactionID        = "PPP_TransactionID"
	FieldStatus                  = "Status"
	FieldProductID               = "productId"
	FieldPaymentMethod           = "payment_method"
	FieldPPPStatus               = "ppp_status"
	FieldMessage                 = "message"
	FieldTransactionType         = "transactionType"
	FieldType                    = "type"
	FieldClientRequestID         = "clientRequestId"
	FieldClientUniqueID          = "clientUniqueId"
	FieldUserPaymentOptionID     = "userPaymentOptionId"
	FieldTransactionID           = "TransactionID"
	FieldAdvanceResponseChecksum = "advanceResponseChecksum"
	FieldReason                  = "Reason"
	FieldReasonCode              = "ReasonCode"
	FieldErrCode                 = "ErrCode"
)

var requiredKeys = []string{
	FieldMerchantID,
	FieldMerchantSiteID,
	FieldTotalAmount,
	FieldCurrency,
	FieldResponseTimeStamp,
	FieldPPPTransactionID,
	FieldStatus,
	FieldProductID,
	FieldPaymentMethod,
	FieldPPPStatus,
	FieldMessage,
	FieldTransactionType,
	FieldType,
	FieldClientRequestID,
	FieldClientUniqueID,
	FieldUserPaymentOptionID,
	FieldAdvanceResponseChecksum,
}

var nonEmptyKeys = []string{
	FieldMerchantID,
	FieldMerchantSiteID,
	FieldTotalAmount,
	FieldCurrency,
	FieldResponseTimeStamp,
	FieldPPPTransactionID,
	FieldStatus,
	FieldPaymentMethod,
	FieldPPPStatus,
	FieldMessage,
	FieldTransactionType,
	FieldType,
	FieldClientRequestID,
	FieldClientUniqueID,
	FieldUserPaymentOptionID,
	FieldAdvanceResponseChecksum,
}

type Options struct {
	MerchantID          string
	MerchantSiteID      string
	MerchantSecretKey   string
	TotalAmount         string
	Currency            string
	ResponseTimeStamp   string
	PPPTransactionID    string
	Status              string
	ProductID           string
	PaymentMethod       string
	PPPStatus           string
	Message             string
	TransactionType     string
	Type                string
	ClientRequestID     string
	ClientUniqueID      string
	UserPaymentOptionID string
	TransactionID       string
	Reason              string
	ReasonCode          string
	ErrCode             string
	Now                 func() time.Time
}

type Payload struct {
	Fields map[string]string
}

func PPPStatusForStatus(status string) (string, error) {
	switch status {
	case StatusPending:
		return PPPStatusPending, nil
	case StatusApproved:
		return PPPStatusOK, nil
	case StatusDeclined:
		return PPPStatusFail, nil
	default:
		return "", fmt.Errorf("unsupported payment status %q", status)
	}
}

func Build(opts Options) (Payload, error) {
	if opts.MerchantSecretKey == "" {
		return Payload{}, errors.New("merchant secret key is required")
	}

	if opts.Status == "" {
		return Payload{}, errors.New("status is required")
	}

	pppStatus := opts.PPPStatus
	if pppStatus == "" {
		var err error
		pppStatus, err = PPPStatusForStatus(opts.Status)
		if err != nil {
			return Payload{}, err
		}
	}

	if opts.ResponseTimeStamp == "" {
		now := time.Now().UTC()
		if opts.Now != nil {
			now = opts.Now().UTC()
		}
		opts.ResponseTimeStamp = FormatResponseTimeStamp(now)
	}

	if opts.PPPTransactionID == "" {
		opts.PPPTransactionID = generatedID("ppp", opts.ResponseTimeStamp)
	}
	if opts.ClientRequestID == "" {
		opts.ClientRequestID = generatedID("req", opts.ResponseTimeStamp)
	}
	if opts.ClientUniqueID == "" {
		opts.ClientUniqueID = generatedID("uniq", opts.ResponseTimeStamp)
	}
	if opts.UserPaymentOptionID == "" {
		opts.UserPaymentOptionID = generatedID("upo", opts.ResponseTimeStamp)
	}
	if opts.Message == "" {
		opts.Message = opts.Status
	}

	fields := map[string]string{
		FieldMerchantID:          opts.MerchantID,
		FieldMerchantSiteID:      opts.MerchantSiteID,
		FieldTotalAmount:         opts.TotalAmount,
		FieldCurrency:            opts.Currency,
		FieldResponseTimeStamp:   opts.ResponseTimeStamp,
		FieldPPPTransactionID:    opts.PPPTransactionID,
		FieldStatus:              opts.Status,
		FieldProductID:           opts.ProductID,
		FieldPaymentMethod:       opts.PaymentMethod,
		FieldPPPStatus:           pppStatus,
		FieldMessage:             opts.Message,
		FieldTransactionType:     opts.TransactionType,
		FieldType:                opts.Type,
		FieldClientRequestID:     opts.ClientRequestID,
		FieldClientUniqueID:      opts.ClientUniqueID,
		FieldUserPaymentOptionID: opts.UserPaymentOptionID,
	}

	setOptional(fields, FieldTransactionID, opts.TransactionID)
	setOptional(fields, FieldReason, opts.Reason)
	setOptional(fields, FieldReasonCode, opts.ReasonCode)
	setOptional(fields, FieldErrCode, opts.ErrCode)

	fields[FieldAdvanceResponseChecksum] = checksum.PaymentAdvanceResponseChecksum(checksum.PaymentFields{
		TotalAmount:       fields[FieldTotalAmount],
		Currency:          fields[FieldCurrency],
		ResponseTimeStamp: fields[FieldResponseTimeStamp],
		PPPTransactionID:  fields[FieldPPPTransactionID],
		Status:            fields[FieldStatus],
		ProductID:         fields[FieldProductID],
		MerchantSecretKey: opts.MerchantSecretKey,
	})

	payload := Payload{Fields: fields}
	if err := payload.Validate(); err != nil {
		return Payload{}, err
	}

	return payload, nil
}

func (p Payload) Values() url.Values {
	values := make(url.Values, len(p.Fields))
	for key, value := range p.Fields {
		values.Set(key, value)
	}
	return values
}

func (p Payload) Encode() string {
	return p.Values().Encode()
}

func ParseEncoded(raw string) (Payload, error) {
	values, err := url.ParseQuery(strings.TrimSpace(raw))
	if err != nil {
		return Payload{}, fmt.Errorf("parse raw URL-encoded payload: %w", err)
	}

	fields := make(map[string]string, len(values))
	for key, vals := range values {
		if len(vals) == 0 {
			fields[key] = ""
			continue
		}
		fields[key] = vals[len(vals)-1]
	}

	payload := Payload{Fields: fields}
	if err := payload.Validate(); err != nil {
		return Payload{}, err
	}

	return payload, nil
}

func RecomputeAdvanceResponseChecksum(payload Payload, merchantSecretKey string) (Payload, error) {
	if strings.TrimSpace(merchantSecretKey) == "" {
		return Payload{}, errors.New("merchant secret key is required")
	}

	next := Payload{Fields: make(map[string]string, len(payload.Fields))}
	for key, value := range payload.Fields {
		next.Fields[key] = value
	}

	next.Fields[FieldAdvanceResponseChecksum] = checksum.PaymentAdvanceResponseChecksum(checksum.PaymentFields{
		TotalAmount:       next.Fields[FieldTotalAmount],
		Currency:          next.Fields[FieldCurrency],
		ResponseTimeStamp: next.Fields[FieldResponseTimeStamp],
		PPPTransactionID:  next.Fields[FieldPPPTransactionID],
		Status:            next.Fields[FieldStatus],
		ProductID:         next.Fields[FieldProductID],
		MerchantSecretKey: merchantSecretKey,
	})

	if err := next.Validate(); err != nil {
		return Payload{}, err
	}

	return next, nil
}

func (p Payload) Validate() error {
	for _, key := range requiredKeys {
		if _, ok := p.Fields[key]; !ok {
			return fmt.Errorf("missing required payment DMN field %q", key)
		}
	}

	for _, key := range nonEmptyKeys {
		if p.Fields[key] == "" {
			return fmt.Errorf("required payment DMN field %q is empty", key)
		}
	}

	return nil
}

func FormatResponseTimeStamp(t time.Time) string {
	return t.UTC().Format("2006-01-02.15:04:05")
}

func generatedID(prefix, seed string) string {
	compactSeed := strings.NewReplacer("-", "", ".", "", ":", "").Replace(seed)
	return prefix + "-" + compactSeed
}

func setOptional(fields map[string]string, key, value string) {
	if value != "" {
		fields[key] = value
	}
}
