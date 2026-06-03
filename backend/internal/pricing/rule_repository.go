package pricing

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminRule struct {
	ID          int64    `json:"id"`
	ProductID   string   `json:"productId"`
	Type        RuleType `json:"ruleType"`
	MinQuantity int      `json:"minQuantity"`
	MaxQuantity *int     `json:"maxQuantity"`
	UnitPrice   *int     `json:"unitPrice"`
	TotalPrice  *int     `json:"totalPrice"`
	Priority    int      `json:"priority"`
	IsActive    bool     `json:"isActive"`
}

type RuleRepository struct {
	db *pgxpool.Pool
}

func NewRuleRepository(db *pgxpool.Pool) *RuleRepository {
	return &RuleRepository{db: db}
}

func (repository *RuleRepository) ListByProduct(ctx context.Context, productID string) ([]AdminRule, error) {
	rows, err := repository.db.Query(ctx, `
		SELECT id, product_id, rule_type, min_quantity, max_quantity, unit_price, total_price, priority, is_active
		FROM product_price_rules
		WHERE product_id = $1
		ORDER BY is_active DESC, priority DESC, min_quantity, id
	`, productID)
	if err != nil {
		return nil, fmt.Errorf("list price rules: %w", err)
	}
	defer rows.Close()

	var rules []AdminRule
	for rows.Next() {
		var rule AdminRule
		if err := rows.Scan(&rule.ID, &rule.ProductID, &rule.Type, &rule.MinQuantity, &rule.MaxQuantity, &rule.UnitPrice, &rule.TotalPrice, &rule.Priority, &rule.IsActive); err != nil {
			return nil, fmt.Errorf("scan price rule: %w", err)
		}
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate price rules: %w", err)
	}
	return rules, nil
}

func (repository *RuleRepository) Create(ctx context.Context, rule AdminRule) (int64, error) {
	var id int64
	err := repository.db.QueryRow(ctx, `
		INSERT INTO product_price_rules (product_id, rule_type, min_quantity, max_quantity, unit_price, total_price, priority, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, rule.ProductID, rule.Type, rule.MinQuantity, rule.MaxQuantity, rule.UnitPrice, rule.TotalPrice, rule.Priority, rule.IsActive).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create price rule: %w", err)
	}
	return id, nil
}

func (repository *RuleRepository) Update(ctx context.Context, rule AdminRule) error {
	result, err := repository.db.Exec(ctx, `
		UPDATE product_price_rules
		SET product_id = $2, rule_type = $3, min_quantity = $4, max_quantity = $5, unit_price = $6, total_price = $7, priority = $8, is_active = $9, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, rule.ID, rule.ProductID, rule.Type, rule.MinQuantity, rule.MaxQuantity, rule.UnitPrice, rule.TotalPrice, rule.Priority, rule.IsActive)
	if err != nil {
		return fmt.Errorf("update price rule: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("price rule not found: %d", rule.ID)
	}
	return nil
}
