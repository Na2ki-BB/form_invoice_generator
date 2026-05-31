package product

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

type Product struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Category      string `json:"category"`
	BaseUnitPrice int    `json:"baseUnitPrice"`
	IsActive      bool   `json:"isActive"`
	SortOrder     int    `json:"sortOrder"`
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (repository *Repository) List(ctx context.Context) ([]Product, error) {
	rows, err := repository.db.Query(ctx, `
		SELECT id, name, description, category, base_unit_price, is_active, sort_order
		FROM products
		ORDER BY sort_order, id
	`)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Category, &product.BaseUnitPrice, &product.IsActive, &product.SortOrder); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, product)
	}
	return products, rows.Err()
}

func (repository *Repository) Create(ctx context.Context, product Product) error {
	_, err := repository.db.Exec(ctx, `
		INSERT INTO products (id, owner_id, name, description, category, base_unit_price, is_active, sort_order)
		VALUES ($1, 1, $2, $3, $4, $5, $6, $7)
	`, product.ID, product.Name, product.Description, product.Category, product.BaseUnitPrice, product.IsActive, product.SortOrder)
	if err != nil {
		return fmt.Errorf("create product: %w", err)
	}
	return nil
}

func (repository *Repository) Update(ctx context.Context, product Product) error {
	result, err := repository.db.Exec(ctx, `
		UPDATE products
		SET name = $2, description = $3, category = $4, base_unit_price = $5, is_active = $6, sort_order = $7, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, product.ID, product.Name, product.Description, product.Category, product.BaseUnitPrice, product.IsActive, product.SortOrder)
	if err != nil {
		return fmt.Errorf("update product: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found: %s", product.ID)
	}
	return nil
}
