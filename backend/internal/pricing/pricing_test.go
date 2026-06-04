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

func TestCalculateWithRules(t *testing.T) {
	product := Product{ID: "test", Name: "テスト商品", UnitPrice: 100}
	two := 2
	three := 3
	ninety := 90
	ninetyFive := 95
	oneHundred := 100
	oneHundredNinety := 190

	tests := []struct {
		name     string
		quantity int
		rules    []Rule
		want     Item
	}{
		{
			name:     "ルールなしは通常価格",
			quantity: 3,
			want:     Item{ProductID: "test", Name: "テスト商品", UnitPrice: 100, Quantity: 3, Amount: 300},
		},
		{
			name:     "固定合計を優先",
			quantity: 2,
			rules: []Rule{
				{Type: RuleTypeTierUnit, MinQuantity: 1, MaxQuantity: &two, UnitPrice: &oneHundred},
				{Type: RuleTypeFixedTotal, MinQuantity: 2, MaxQuantity: &two, TotalPrice: &oneHundredNinety},
			},
			want: Item{ProductID: "test", Name: "テスト商品", UnitPrice: 95, Quantity: 2, Amount: 190},
		},
		{
			name:     "固定合計の後は段階単価を使用",
			quantity: 3,
			rules: []Rule{
				{Type: RuleTypeFixedTotal, MinQuantity: 2, MaxQuantity: &two, TotalPrice: &oneHundredNinety},
				{Type: RuleTypeTierUnit, MinQuantity: 2, MaxQuantity: nil, UnitPrice: &ninetyFive},
			},
			want: Item{ProductID: "test", Name: "テスト商品", UnitPrice: 95, Quantity: 3, Amount: 285},
		},
		{
			name:     "段階単価を使用",
			quantity: 3,
			rules: []Rule{
				{Type: RuleTypeTierUnit, MinQuantity: 3, MaxQuantity: nil, UnitPrice: &ninety},
			},
			want: Item{ProductID: "test", Name: "テスト商品", UnitPrice: 90, Quantity: 3, Amount: 270},
		},
		{
			name:     "同種ルールはpriorityが高いものを使用",
			quantity: 3,
			rules: []Rule{
				{Type: RuleTypeTierUnit, MinQuantity: 3, MaxQuantity: &three, UnitPrice: &oneHundred, Priority: 1},
				{Type: RuleTypeTierUnit, MinQuantity: 3, MaxQuantity: &three, UnitPrice: &ninety, Priority: 2},
			},
			want: Item{ProductID: "test", Name: "テスト商品", UnitPrice: 90, Quantity: 3, Amount: 270},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := CalculateWithRules(product, test.quantity, test.rules)
			if err != nil {
				t.Fatalf("CalculateWithRules() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("CalculateWithRules() = %#v, want %#v", got, test.want)
			}
		})
	}
}
