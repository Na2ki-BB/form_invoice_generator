package pricing

import "testing"

func TestCalculate(t *testing.T) {
	tests := []struct {
		name      string
		productID string
		quantity  int
		want      Item
	}{
		{
			name:      "祈祷Aを2件",
			productID: "prayer-a",
			quantity:  2,
			want:      Item{ProductID: "prayer-a", Name: "祈祷A", UnitPrice: 5000, Quantity: 2, Amount: 10000},
		},
		{
			name:      "御札を3件",
			productID: "ofuda",
			quantity:  3,
			want:      Item{ProductID: "ofuda", Name: "御札", UnitPrice: 1000, Quantity: 3, Amount: 3000},
		},
		{
			name:      "お守りを0件",
			productID: "omamori",
			quantity:  0,
			want:      Item{ProductID: "omamori", Name: "お守り", UnitPrice: 800, Quantity: 0, Amount: 0},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Calculate(test.productID, test.quantity)
			if err != nil {
				t.Fatalf("Calculate() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("Calculate() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestCalculateRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name      string
		productID string
		quantity  int
	}{
		{name: "存在しない商品", productID: "unknown", quantity: 1},
		{name: "数量がマイナス", productID: "ofuda", quantity: -1},
		{name: "数量が上限超過", productID: "ofuda", quantity: 11},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Calculate(test.productID, test.quantity); err == nil {
				t.Fatal("Calculate() error = nil, want error")
			}
		})
	}
}
