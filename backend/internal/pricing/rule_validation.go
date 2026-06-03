package pricing

import "fmt"

func ValidateAdminRule(rule AdminRule) error {
	if rule.ProductID == "" {
		return fmt.Errorf("productId is required")
	}
	if rule.MinQuantity < 1 {
		return fmt.Errorf("minQuantity must be at least 1")
	}
	if rule.MaxQuantity != nil && *rule.MaxQuantity < rule.MinQuantity {
		return fmt.Errorf("maxQuantity must be greater than or equal to minQuantity")
	}

	switch rule.Type {
	case RuleTypeFixedTotal:
		if rule.TotalPrice == nil || *rule.TotalPrice < 0 || rule.UnitPrice != nil {
			return fmt.Errorf("fixed_total requires totalPrice only")
		}
		if rule.MaxQuantity == nil || *rule.MaxQuantity != rule.MinQuantity {
			return fmt.Errorf("fixed_total must target one exact quantity")
		}
		if *rule.TotalPrice%rule.MinQuantity != 0 {
			return fmt.Errorf("fixed_total must be divisible by quantity")
		}
	case RuleTypeTierUnit:
		if rule.UnitPrice == nil || *rule.UnitPrice < 0 || rule.TotalPrice != nil {
			return fmt.Errorf("tier_unit requires unitPrice only")
		}
	default:
		return fmt.Errorf("unknown rule type: %s", rule.Type)
	}
	return nil
}
