package pricing

import "testing"

func TestValidateAdminRule(t *testing.T) {
	two := 2
	unitPrice := 900
	totalPrice := 1900
	undivisibleTotal := 1000

	tests := []struct {
		name      string
		rule      AdminRule
		wantError bool
	}{
		{name: "valid fixed total", rule: AdminRule{ProductID: "ofuda", Type: RuleTypeFixedTotal, MinQuantity: 2, MaxQuantity: &two, TotalPrice: &totalPrice}},
		{name: "fixed total must be divisible", rule: AdminRule{ProductID: "ofuda", Type: RuleTypeFixedTotal, MinQuantity: 3, MaxQuantity: intPtr(3), TotalPrice: &undivisibleTotal}, wantError: true},
		{name: "valid tier unit", rule: AdminRule{ProductID: "ofuda", Type: RuleTypeTierUnit, MinQuantity: 3, UnitPrice: &unitPrice}},
		{name: "product required", rule: AdminRule{Type: RuleTypeTierUnit, MinQuantity: 1, UnitPrice: &unitPrice}, wantError: true},
		{name: "unknown type", rule: AdminRule{ProductID: "ofuda", Type: "unknown", MinQuantity: 1, UnitPrice: &unitPrice}, wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateAdminRule(test.rule)
			if (err != nil) != test.wantError {
				t.Fatalf("ValidateAdminRule() error = %v, wantError = %v", err, test.wantError)
			}
		})
	}
}

func intPtr(value int) *int {
	return &value
}
