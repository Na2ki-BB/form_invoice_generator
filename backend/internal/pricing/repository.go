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
	`, product.ID).Scan(&product.ID, &product.Name, &product.UnitPrice)
	if err != nil {
		return Item{}, fmt.Errorf("find pricing product: %w", err)
	}
	return repository.calculateProduct(ctx, product, quantity)
}

func (repository *Repository) CalculateForForm(ctx context.Context, publicSlug string, productID string, quantity int) (Item, error) {
	var product Product
	var minQuantity int
	var maxQuantity int
	err := repository.db.QueryRow(ctx, `
		SELECT p.id, p.name, p.base_unit_price, fp.min_quantity, fp.max_quantity
		FROM forms f
		JOIN form_products fp ON fp.form_id = f.id
		JOIN products p ON p.id = fp.product_id
		WHERE f.public_slug = $1 AND f.is_active = TRUE AND p.id = $2 AND p.is_active = TRUE
	`, publicSlug, productID).Scan(&product.ID, &product.Name, &product.UnitPrice, &minQuantity, &maxQuantity)
	if err != nil {
		return Item{}, fmt.Errorf("find form pricing product: %w", err)
	}
	if quantity < minQuantity || quantity > maxQuantity {
		return Item{}, fmt.Errorf("quantity must be between %d and %d: %d", minQuantity, maxQuantity, quantity)
	}
	return repository.calculateProduct(ctx, product, quantity)
}

func (repository *Repository) calculateProduct(ctx context.Context, product Product, quantity int) (Item, error) {
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
