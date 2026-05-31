package submission

import (
	"context"
	"fmt"

	"form-invoice-generator/backend/internal/pricing"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

type Submission struct {
	CustomerName  string
	CustomerKana  string
	CustomerEmail string
	CustomerPhone string
	PostalCode    string
	Address       string
	Note          string
	TotalAmount   int
	Items         []pricing.Item
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (repository *Repository) Create(ctx context.Context, submission Submission) (int64, error) {
	transaction, err := repository.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin submission transaction: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	var submissionID int64
	err = transaction.QueryRow(ctx, `
		INSERT INTO submissions (
			customer_name,
			customer_kana,
			customer_email,
			customer_phone,
			postal_code,
			address,
			note,
			total_amount
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`,
		submission.CustomerName,
		submission.CustomerKana,
		submission.CustomerEmail,
		submission.CustomerPhone,
		submission.PostalCode,
		submission.Address,
		submission.Note,
		submission.TotalAmount,
	).Scan(&submissionID)
	if err != nil {
		return 0, fmt.Errorf("insert submission: %w", err)
	}

	for _, item := range submission.Items {
		_, err := transaction.Exec(ctx, `
			INSERT INTO submission_items (
				submission_id,
				product_id,
				product_name_snapshot,
				quantity,
				unit_price_snapshot,
				total_amount_snapshot
			) VALUES ($1, $2, $3, $4, $5, $6)
		`,
			submissionID,
			item.ProductID,
			item.Name,
			item.Quantity,
			item.UnitPrice,
			item.Amount,
		)
		if err != nil {
			return 0, fmt.Errorf("insert submission item: %w", err)
		}
	}

	if err := transaction.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit submission transaction: %w", err)
	}
	return submissionID, nil
}
