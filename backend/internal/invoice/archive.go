package invoice

import (
	"archive/zip"
	"bytes"
	"fmt"

	"form-invoice-generator/backend/internal/submission"
)

func GenerateArchive(submissions []submission.Detail) ([]byte, error) {
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)

	for _, detail := range submissions {
		generated, err := Generate(Data{
			InvoiceNumber: fmt.Sprintf("INV-TEMP-%04d", detail.ID),
			InvoiceDate:   detail.SubmittedAt,
			CustomerName:  detail.CustomerName,
			PostalCode:    detail.PostalCode,
			Address:       detail.Address,
			Note:          detail.Note,
			Items:         detail.Items,
		})
		if err != nil {
			_ = archive.Close()
			return nil, fmt.Errorf("generate invoice for submission %d: %w", detail.ID, err)
		}

		file, err := archive.Create(fmt.Sprintf("invoice_%04d.xlsx", detail.ID))
		if err != nil {
			_ = archive.Close()
			return nil, fmt.Errorf("create invoice archive entry: %w", err)
		}
		if _, err := file.Write(generated); err != nil {
			_ = archive.Close()
			return nil, fmt.Errorf("write invoice archive entry: %w", err)
		}
	}

	if err := archive.Close(); err != nil {
		return nil, fmt.Errorf("close invoice archive: %w", err)
	}
	return buffer.Bytes(), nil
}
