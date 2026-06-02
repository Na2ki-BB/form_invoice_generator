package form

import "testing"

func TestValidateProductCount(t *testing.T) {
	tests := []struct {
		name       string
		productIDs []string
		wantError  bool
	}{
		{name: "five products", productIDs: []string{"1", "2", "3", "4", "5"}},
		{name: "six products", productIDs: []string{"1", "2", "3", "4", "5", "6"}, wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateProductCount(AdminForm{ProductIDs: test.productIDs}, 5)
			if (err != nil) != test.wantError {
				t.Fatalf("ValidateProductCount() error = %v, wantError = %v", err, test.wantError)
			}
		})
	}
}
