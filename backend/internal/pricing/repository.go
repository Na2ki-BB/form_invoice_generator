package pricing

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (repository *Repository) Calculate(ctx context.Context, productID string, quantity int) (Item, error) {
	var product Product
	err := repository.db.QueryRow(ctx, `
		SELECT id, name, base_unit_price
		FROM products
		WHERE id = $1 AND is_active = TRUE
	`, productID).Scan(&product.ID, &product.Name, &product.UnitPrice)
	if err != nil {
		return Item{}, fmt.Errorf("find pricing product: %w", err)
	}

	rows, err := repository.db.Query(ctx, `
		SELECT rule_type, min_quantity, max_quantity, unit_price, total_price, priority
		FROM product_price_rules
		WHERE product_id = $1 AND is_active = TRUE
		ORDER BY priority DESC, id
	`, productID)
	if err != nil {
		return Item{}, fmt.Errorf("find pricing rules: %w", err)
	}
	defer rows.Close()

	var rules []Rule
	for rows.Next() {
		var rule Rule
		if err := rows.Scan(
			&rule.Type,
			&rule.MinQuantity,
			&rule.MaxQuantity,
			&rule.UnitPrice,
			&rule.TotalPrice,
			&rule.Priority,
		); err != nil {
			return Item{}, fmt.Errorf("scan pricing rule: %w", err)
		}
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return Item{}, fmt.Errorf("iterate pricing rules: %w", err)
	}

	return CalculateWithRules(product, quantity, rules)
}
