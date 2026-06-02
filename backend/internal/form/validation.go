package form

import "fmt"

func ValidateProductCount(form AdminForm, maxProducts int) error {
	if len(form.ProductIDs) > maxProducts {
		return fmt.Errorf("form supports up to %d products", maxProducts)
	}
	return nil
}
