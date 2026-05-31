package pricing

import "fmt"

type Product struct {
	ID        string
	Name      string
	UnitPrice int
}

type RuleType string

const (
	RuleTypeFixedTotal RuleType = "fixed_total"
	RuleTypeTierUnit   RuleType = "tier_unit"
)

type Rule struct {
	Type        RuleType
	MinQuantity int
	MaxQuantity *int
	UnitPrice   *int
	TotalPrice  *int
	Priority    int
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
	return CalculateWithRules(product, quantity, nil)
}

func CalculateWithRules(product Product, quantity int, rules []Rule) (Item, error) {
	if quantity < 0 || quantity > 10 {
		return Item{}, fmt.Errorf("quantity must be between 0 and 10: %d", quantity)
	}

	unitPrice := product.UnitPrice
	amount := product.UnitPrice * quantity
	if quantity > 0 {
		if rule, ok := findFixedTotalRule(rules, quantity); ok {
			amount = *rule.TotalPrice
			unitPrice = amount / quantity
		} else if rule, ok := findTierUnitRule(rules, quantity); ok {
			unitPrice = *rule.UnitPrice
			amount = unitPrice * quantity
		}
	}

	return Item{
		ProductID: product.ID,
		Name:      product.Name,
		UnitPrice: unitPrice,
		Quantity:  quantity,
		Amount:    amount,
	}, nil
}

func findFixedTotalRule(rules []Rule, quantity int) (Rule, bool) {
	return findBestRule(rules, RuleTypeFixedTotal, quantity)
}

func findTierUnitRule(rules []Rule, quantity int) (Rule, bool) {
	return findBestRule(rules, RuleTypeTierUnit, quantity)
}

func findBestRule(rules []Rule, ruleType RuleType, quantity int) (Rule, bool) {
	var best Rule
	found := false
	for _, rule := range rules {
		if rule.Type != ruleType || quantity < rule.MinQuantity {
			continue
		}
		if rule.MaxQuantity != nil && quantity > *rule.MaxQuantity {
			continue
		}
		if !found || rule.Priority > best.Priority {
			best = rule
			found = true
		}
	}
	return best, found
}
