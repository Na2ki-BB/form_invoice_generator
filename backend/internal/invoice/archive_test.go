package invoice

import (
	"archive/zip"
	"bytes"
	"testing"
	"time"

	"form-invoice-generator/backend/internal/pricing"
	"form-invoice-generator/backend/internal/submission"
)

func TestGenerateArchive(t *testing.T) {
	generated, err := GenerateArchive([]submission.Detail{
		{
			ID:            12,
			InvoiceNumber: "INV-202606-000012",
			CustomerName:  "山田 太郎",
			SubmittedAt:   time.Date(2026, time.June, 1, 0, 0, 0, 0, time.Local),
			Items:         []pricing.Item{{ProductID: "prayer-a", Name: "祈祷A", UnitPrice: 5000, Quantity: 1, Amount: 5000}},
		},
		{
			ID:            34,
			InvoiceNumber: "INV-202606-000034",
			CustomerName:  "鈴木 花子",
			SubmittedAt:   time.Date(2026, time.June, 2, 0, 0, 0, 0, time.Local),
			Items:         []pricing.Item{{ProductID: "ofuda", Name: "御札", UnitPrice: 950, Quantity: 2, Amount: 1900}},
		},
	})
	if err != nil {
		t.Fatalf("GenerateArchive() error = %v", err)
	}

	archive, err := zip.NewReader(bytes.NewReader(generated), int64(len(generated)))
	if err != nil {
		t.Fatalf("open generated archive: %v", err)
	}
	if len(archive.File) != 2 {
		t.Fatalf("archive file count = %d, want 2", len(archive.File))
	}

	wantNames := []string{"invoice_0012.xlsx", "invoice_0034.xlsx"}
	for index, file := range archive.File {
		if file.Name != wantNames[index] {
			t.Fatalf("archive file[%d] = %q, want %q", index, file.Name, wantNames[index])
		}
		if file.UncompressedSize64 == 0 {
			t.Fatalf("archive file[%d] is empty", index)
		}
	}
}
