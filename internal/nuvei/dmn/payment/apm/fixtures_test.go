package apm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/parkerscobey/nuvei-dmn-simulator/internal/nuvei/dmn/payment"
)

func TestFixturePayloadsParse(t *testing.T) {
	tests := []struct {
		name              string
		file              string
		wantPaymentMethod string
	}{
		{name: "pix", file: "pix_payload.txt", wantPaymentMethod: PixPaymentMethod},
		{name: "boleto", file: "boleto_payload.txt", wantPaymentMethod: BoletoPaymentMethod},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("testdata", tt.file))
			if err != nil {
				t.Fatalf("ReadFile(%q) error = %v", tt.file, err)
			}

			p, err := payment.ParseEncoded(string(raw))
			if err != nil {
				t.Fatalf("ParseEncoded(%q) error = %v", tt.file, err)
			}

			assertField(t, p, payment.FieldPaymentMethod, tt.wantPaymentMethod)
		})
	}
}
