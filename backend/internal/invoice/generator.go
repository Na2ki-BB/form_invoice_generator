package invoice

import (
	"bytes"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"form-invoice-generator/backend/internal/pricing"
	"github.com/xuri/excelize/v2"
)

const sheetName = "請求書"

const MaxItems = 5

type Data struct {
	InvoiceNumber string
	InvoiceDate   time.Time
	CustomerName  string
	PostalCode    string
	Address       string
	Note          string
	Items         []pricing.Item
}

func Generate(data Data) ([]byte, error) {
	templatePath, err := templateFilePath()
	if err != nil {
		return nil, err
	}

	file, err := excelize.OpenFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("open invoice template: %w", err)
	}
	defer func() { _ = file.Close() }()

	if err := setValue(file, "C3", data.InvoiceNumber); err != nil {
		return nil, err
	}
	if err := setValue(file, "E3", data.InvoiceDate.Format("2006-01-02")); err != nil {
		return nil, err
	}
	if err := setValue(file, "C5", data.CustomerName); err != nil {
		return nil, err
	}
	if err := setValue(file, "C6", data.PostalCode); err != nil {
		return nil, err
	}
	if err := setValue(file, "C7", data.Address); err != nil {
		return nil, err
	}
	if err := setValue(file, "B19", data.Note); err != nil {
		return nil, err
	}

	totalAmount := 0
	for index, item := range data.Items {
		row := 10 + index
		if index >= MaxItems {
			return nil, fmt.Errorf("invoice supports up to %d items", MaxItems)
		}
		if err := setValue(file, fmt.Sprintf("B%d", row), item.Name); err != nil {
			return nil, err
		}
		if err := setValue(file, fmt.Sprintf("C%d", row), item.Quantity); err != nil {
			return nil, err
		}
		if err := setValue(file, fmt.Sprintf("D%d", row), item.UnitPrice); err != nil {
			return nil, err
		}
		if err := setValue(file, fmt.Sprintf("E%d", row), item.Amount); err != nil {
			return nil, err
		}
		totalAmount += item.Amount
	}
	if err := setValue(file, "E16", totalAmount); err != nil {
		return nil, err
	}

	buffer, err := file.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("write invoice: %w", err)
	}
	return bytes.Clone(buffer.Bytes()), nil
}

func setValue(file *excelize.File, cell string, value any) error {
	if err := file.SetCellValue(sheetName, cell, value); err != nil {
		return fmt.Errorf("set invoice cell %s: %w", cell, err)
	}
	return nil
}

func templateFilePath() (string, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("find invoice template path")
	}
	return filepath.Join(filepath.Dir(currentFile), "..", "..", "templates", "template_invoice.xlsx"), nil
}
