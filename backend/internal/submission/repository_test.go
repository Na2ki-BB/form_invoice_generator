package submission

import (
	"testing"
	"time"
)

func TestFormatInvoiceNumber(t *testing.T) {
	submittedAt := time.Date(2026, time.June, 2, 21, 50, 0, 0, time.FixedZone("JST", 9*60*60))
	got := FormatInvoiceNumber(10, submittedAt)
	want := "INV-202606-000010"
	if got != want {
		t.Fatalf("FormatInvoiceNumber() = %q, want %q", got, want)
	}
}
