package form

import (
	"fmt"
	"regexp"
	"strings"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

func Validate(form AdminForm, maxProducts int) error {
	if strings.TrimSpace(form.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if !slugPattern.MatchString(form.PublicSlug) {
		return fmt.Errorf("publicSlug must use lowercase letters, numbers, and hyphens")
	}
	if len(form.ProductIDs) > maxProducts {
		return fmt.Errorf("form supports up to %d products", maxProducts)
	}
	seen := map[string]bool{}
	for _, productID := range form.ProductIDs {
		if strings.TrimSpace(productID) == "" {
			return fmt.Errorf("product id is required")
		}
		if seen[productID] {
			return fmt.Errorf("duplicate product id: %s", productID)
		}
		seen[productID] = true
	}
	return nil
}

func ValidateProductCount(form AdminForm, maxProducts int) error {
	return Validate(form, maxProducts)
}
