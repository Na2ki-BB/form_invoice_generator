package pricing

import "fmt"

type Product struct {
	ID        string
	Name      string
	UnitPrice int
}

type Item struct {
	ProductID string
	Name      string
	UnitPrice int
	Quantity  int
	Amount    int
}

var products = map[string]Product{
	"prayer-a": {ID: "prayer-a", Name: "祈祷A", UnitPrice: 5000},
	"ofuda":    {ID: "ofuda", Name: "御札", UnitPrice: 1000},
	"omamori":  {ID: "omamori", Name: "お守り", UnitPrice: 800},
}

func Calculate(productID string, quantity int) (Item, error) {
	product, ok := products[productID]
	if !ok {
		return Item{}, fmt.Errorf("unknown product: %s", productID)
	}
	if quantity < 0 || quantity > 10 {
		return Item{}, fmt.Errorf("quantity must be between 0 and 10: %d", quantity)
	}

	return Item{
		ProductID: product.ID,
		Name:      product.Name,
		UnitPrice: product.UnitPrice,
		Quantity:  quantity,
		Amount:    product.UnitPrice * quantity,
	}, nil
}
