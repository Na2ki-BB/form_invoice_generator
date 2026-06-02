package submission

import (
	"testing"

	"form-invoice-generator/backend/internal/pricing"
)

func TestValidate(t *testing.T) {
	valid := Submission{
		CustomerName:  "山田 太郎",
		PostalCode:    "100-0001",
		Address:       "東京都千代田区",
		CustomerPhone: "03-0000-0000",
		Items:         []pricing.Item{{ProductID: "prayer-a", Quantity: 1}},
	}

	tests := []struct {
		name       string
		submission Submission
		wantError  bool
	}{
		{name: "valid", submission: valid},
		{name: "customer name is required", submission: withCustomerName(valid, " "), wantError: true},
		{name: "postal code is required", submission: withPostalCode(valid, ""), wantError: true},
		{name: "address is required", submission: withAddress(valid, "\t"), wantError: true},
		{name: "phone is required", submission: withCustomerPhone(valid, ""), wantError: true},
		{name: "item is required", submission: withItems(valid, nil), wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := Validate(test.submission)
			if (err != nil) != test.wantError {
				t.Fatalf("Validate() error = %v, wantError = %v", err, test.wantError)
			}
		})
	}
}

func withCustomerName(submission Submission, value string) Submission {
	submission.CustomerName = value
	return submission
}

func withPostalCode(submission Submission, value string) Submission {
	submission.PostalCode = value
	return submission
}

func withAddress(submission Submission, value string) Submission {
	submission.Address = value
	return submission
}

func withCustomerPhone(submission Submission, value string) Submission {
	submission.CustomerPhone = value
	return submission
}

func withItems(submission Submission, items []pricing.Item) Submission {
	submission.Items = items
	return submission
}
