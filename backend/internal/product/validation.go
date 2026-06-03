package product

import (
	"fmt"
	"strings"
)

func Validate(product Product) error {
	if strings.TrimSpace(product.ID) == "" {
		return fmt.Errorf("product id is required")
	}
	if strings.TrimSpace(product.Name) == "" {
		return fmt.Errorf("product name is required")
	}
	if strings.TrimSpace(product.Category) == "" {
		return fmt.Errorf("category is required")
	}
	if product.BaseUnitPrice < 0 {
		return fmt.Errorf("base unit price must be greater than or equal to 0")
	}
	return nil
}
