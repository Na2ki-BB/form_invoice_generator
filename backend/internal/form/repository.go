package form

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("form not found")

type Repository struct {
	db *pgxpool.Pool
}

type Form struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	PublicSlug  string    `json:"publicSlug"`
	Products    []Product `json:"products"`
}

type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	UnitPrice   int    `json:"unitPrice"`
	MinQuantity int    `json:"minQuantity"`
	MaxQuantity int    `json:"maxQuantity"`
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (repository *Repository) FindIDBySlug(ctx context.Context, slug string) (int64, error) {
	var id int64
	err := repository.db.QueryRow(ctx, `
		SELECT id
		FROM forms
		WHERE public_slug = $1 AND is_active = TRUE
	`, slug).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("find form id: %w", err)
	}
	return id, nil
}

func (repository *Repository) FindBySlug(ctx context.Context, slug string) (Form, error) {
	var result Form
	err := repository.db.QueryRow(ctx, `
		SELECT title, description, public_slug
		FROM forms
		WHERE public_slug = $1 AND is_active = TRUE
	`, slug).Scan(&result.Title, &result.Description, &result.PublicSlug)
	if errors.Is(err, pgx.ErrNoRows) {
		return Form{}, ErrNotFound
	}
	if err != nil {
		return Form{}, fmt.Errorf("find form: %w", err)
	}

	rows, err := repository.db.Query(ctx, `
		SELECT p.id, p.name, p.description, p.base_unit_price, fp.min_quantity, fp.max_quantity
		FROM form_products fp
		JOIN products p ON p.id = fp.product_id
		JOIN forms f ON f.id = fp.form_id
		WHERE f.public_slug = $1 AND p.is_active = TRUE
		ORDER BY fp.sort_order, p.sort_order, p.id
	`, slug)
	if err != nil {
		return Form{}, fmt.Errorf("find form products: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var product Product
		if err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.UnitPrice,
			&product.MinQuantity,
			&product.MaxQuantity,
		); err != nil {
			return Form{}, fmt.Errorf("scan form product: %w", err)
		}
		result.Products = append(result.Products, product)
	}
	if err := rows.Err(); err != nil {
		return Form{}, fmt.Errorf("iterate form products: %w", err)
	}
	return result, nil
}

type AdminForm struct {
	ID          int64    `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	PublicSlug  string   `json:"publicSlug"`
	IsActive    bool     `json:"isActive"`
	ProductIDs  []string `json:"productIds"`
}

func (repository *Repository) List(ctx context.Context) ([]AdminForm, error) {
	rows, err := repository.db.Query(ctx, `
		SELECT f.id, f.title, f.description, f.public_slug, f.is_active, COALESCE(array_agg(fp.product_id ORDER BY fp.sort_order) FILTER (WHERE fp.product_id IS NOT NULL), '{}')
		FROM forms f
		LEFT JOIN form_products fp ON fp.form_id = f.id
		GROUP BY f.id
		ORDER BY f.id
	`)
	if err != nil {
		return nil, fmt.Errorf("list forms: %w", err)
	}
	defer rows.Close()

	var forms []AdminForm
	for rows.Next() {
		var form AdminForm
		if err := rows.Scan(&form.ID, &form.Title, &form.Description, &form.PublicSlug, &form.IsActive, &form.ProductIDs); err != nil {
			return nil, fmt.Errorf("scan form: %w", err)
		}
		forms = append(forms, form)
	}
	return forms, rows.Err()
}

func (repository *Repository) Create(ctx context.Context, form AdminForm) (int64, error) {
	transaction, err := repository.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin form transaction: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	var formID int64
	err = transaction.QueryRow(ctx, `
		INSERT INTO forms (owner_id, title, description, public_slug, is_active)
		VALUES (1, $1, $2, $3, $4)
		RETURNING id
	`, form.Title, form.Description, form.PublicSlug, form.IsActive).Scan(&formID)
	if err != nil {
		return 0, fmt.Errorf("create form: %w", err)
	}
	if err := replaceProducts(ctx, transaction, formID, form.ProductIDs); err != nil {
		return 0, err
	}
	if err := transaction.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit form transaction: %w", err)
	}
	return formID, nil
}

func (repository *Repository) Update(ctx context.Context, form AdminForm) error {
	transaction, err := repository.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin form transaction: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	result, err := transaction.Exec(ctx, `
		UPDATE forms SET title = $2, description = $3, public_slug = $4, is_active = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, form.ID, form.Title, form.Description, form.PublicSlug, form.IsActive)
	if err != nil {
		return fmt.Errorf("update form: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("form not found: %d", form.ID)
	}
	if err := replaceProducts(ctx, transaction, form.ID, form.ProductIDs); err != nil {
		return err
	}
	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit form transaction: %w", err)
	}
	return nil
}

func replaceProducts(ctx context.Context, transaction pgx.Tx, formID int64, productIDs []string) error {
	if _, err := transaction.Exec(ctx, `DELETE FROM form_products WHERE form_id = $1`, formID); err != nil {
		return fmt.Errorf("delete form products: %w", err)
	}
	for index, productID := range productIDs {
		if _, err := transaction.Exec(ctx, `
			INSERT INTO form_products (form_id, product_id, sort_order, min_quantity, max_quantity)
			VALUES ($1, $2, $3, 0, 10)
		`, formID, productID, (index+1)*10); err != nil {
			return fmt.Errorf("insert form product: %w", err)
		}
	}
	return nil
}
