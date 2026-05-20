package checksum

import (
	"crypto/sha256"
	"encoding/hex"
)

// PaymentFields are the fields used for a payment DMN advanceResponseChecksum.
type PaymentFields struct {
	TotalAmount       string
	Currency          string
	ResponseTimeStamp string
	PPPTransactionID  string
	Status            string
	ProductID         string
	MerchantSecretKey string
}

// PaymentSourceString returns the source string expected by the simulator's
// target receivers for payment DMN checksum validation.
func PaymentSourceString(fields PaymentFields) string {
	return fields.TotalAmount +
		fields.Currency +
		fields.ResponseTimeStamp +
		fields.PPPTransactionID +
		fields.Status +
		fields.ProductID +
		fields.MerchantSecretKey
}

// SHA256Hex returns a lowercase SHA-256 hex digest for source.
func SHA256Hex(source string) string {
	sum := sha256.Sum256([]byte(source))
	return hex.EncodeToString(sum[:])
}

// PaymentAdvanceResponseChecksum returns the SHA-256 advanceResponseChecksum for
// a payment DMN.
func PaymentAdvanceResponseChecksum(fields PaymentFields) string {
	return SHA256Hex(PaymentSourceString(fields))
}
