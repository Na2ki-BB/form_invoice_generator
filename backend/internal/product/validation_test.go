package product

import "testing"

func TestValidate(t *testing.T) {
	valid := Product{ID: "test", Name: "テスト", Category: "その他", BaseUnitPrice: 100}
	tests := []struct {
		name      string
		product   Product
		wantError bool
	}{
		{name: "valid", product: valid},
		{name: "blank id", product: Product{Name: "テスト", Category: "その他"}, wantError: true},
		{name: "blank name", product: Product{ID: "test", Name: " ", Category: "その他"}, wantError: true},
		{name: "blank category", product: Product{ID: "test", Name: "テスト", Category: ""}, wantError: true},
		{name: "negative price", product: Product{ID: "test", Name: "テスト", Category: "その他", BaseUnitPrice: -1}, wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := Validate(test.product)
			if (err != nil) != test.wantError {
				t.Fatalf("Validate() error = %v, wantError = %v", err, test.wantError)
			}
		})
	}
}
