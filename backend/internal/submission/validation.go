package submission

import (
	"fmt"
	"strings"
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
	if len(submission.Items) == 0 {
		return fmt.Errorf("at least one item is required")
	}
	if len(submission.Items) > maxItems {
		return fmt.Errorf("submission supports up to %d items", maxItems)
	}
	return nil
}
