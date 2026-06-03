package submission

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	maxCustomerNameLength  = 100
	maxCustomerKanaLength  = 100
	maxPostalCodeLength    = 20
	maxAddressLength       = 300
	maxCustomerPhoneLength = 50
	maxCustomerEmailLength = 200
	maxNoteLength          = 1000
)

func Validate(submission Submission, maxItems int) error {
	requiredFields := []struct {
		name  string
		value string
	}{
		{name: "customerName", value: submission.CustomerName},
		{name: "postalCode", value: submission.PostalCode},
		{name: "address", value: submission.Address},
		{name: "phone", value: submission.CustomerPhone},
	}
	for _, field := range requiredFields {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("%s is required", field.name)
		}
	}

	lengthFields := []struct {
		name      string
		value     string
		maxLength int
	}{
		{name: "customerName", value: submission.CustomerName, maxLength: maxCustomerNameLength},
		{name: "customerKana", value: submission.CustomerKana, maxLength: maxCustomerKanaLength},
		{name: "postalCode", value: submission.PostalCode, maxLength: maxPostalCodeLength},
		{name: "address", value: submission.Address, maxLength: maxAddressLength},
		{name: "phone", value: submission.CustomerPhone, maxLength: maxCustomerPhoneLength},
		{name: "email", value: submission.CustomerEmail, maxLength: maxCustomerEmailLength},
		{name: "note", value: submission.Note, maxLength: maxNoteLength},
	}
	for _, field := range lengthFields {
		if utf8.RuneCountInString(field.value) > field.maxLength {
			return fmt.Errorf("%s must be %d characters or fewer", field.name, field.maxLength)
		}
	}

	if len(submission.Items) == 0 {
		return fmt.Errorf("at least one item is required")
	}
	if len(submission.Items) > maxItems {
		return fmt.Errorf("submission supports up to %d items", maxItems)
	}
	return nil
}
