package invoice

import (
	"bytes"
	"testing"
	"time"

	"form-invoice-generator/backend/internal/pricing"
	"github.com/xuri/excelize/v2"
)

func TestGenerate(t *testing.T) {
	generated, err := Generate(Data{
		InvoiceNumber: "INV-TEST-0001",
		InvoiceDate:   time.Date(2026, time.May, 31, 0, 0, 0, 0, time.Local),
		CustomerName:  "山田 太郎",
		PostalCode:    "100-0001",
		Address:       "東京都千代田区千代田1-1",
		Note:          "動作確認用",
		Items: []pricing.Item{
			{ProductID: "prayer-a", Name: "祈祷A", UnitPrice: 5000, Quantity: 2, Amount: 10000},
			{ProductID: "omamori", Name: "お守り", UnitPrice: 800, Quantity: 1, Amount: 800},
		},
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	file, err := excelize.OpenReader(bytes.NewReader(generated))
	if err != nil {
		t.Fatalf("open generated invoice: %v", err)
	}
	defer func() { _ = file.Close() }()

	assertCellValue(t, file, "C3", "INV-TEST-0001")
	assertCellValue(t, file, "E3", "2026-05-31")
	assertCellValue(t, file, "C5", "山田 太郎")
	assertCellValue(t, file, "C6", "100-0001")
	assertCellValue(t, file, "C7", "東京都千代田区千代田1-1")
	assertCellValue(t, file, "B10", "祈祷A")
	assertCellValue(t, file, "C10", "2")
	assertCellValue(t, file, "D10", "5000")
	assertCellValue(t, file, "E10", "10000")
	assertCellValue(t, file, "B11", "お守り")
	assertCellValue(t, file, "E11", "800")
	assertCellValue(t, file, "E16", "10800")
	assertCellValue(t, file, "B19", "動作確認用")
}

func assertCellValue(t *testing.T, file *excelize.File, cell string, want string) {
	t.Helper()

	got, err := file.GetCellValue(sheetName, cell)
	if err != nil {
		t.Fatalf("GetCellValue(%s) error = %v", cell, err)
	}
	if got != want {
		t.Fatalf("GetCellValue(%s) = %q, want %q", cell, got, want)
	}
}
