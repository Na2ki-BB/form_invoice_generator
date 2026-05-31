package submission

import (
	"context"
	"fmt"
	"time"

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

type Summary struct {
	ID            int64     `json:"id"`
	CustomerName  string    `json:"customerName"`
	CustomerPhone string    `json:"customerPhone"`
	TotalAmount   int       `json:"totalAmount"`
	SubmittedAt   time.Time `json:"submittedAt"`
	Status        string    `json:"status"`
}

func (repository *Repository) ListByMonth(ctx context.Context, month time.Time) ([]Summary, error) {
	monthStart := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)

	rows, err := repository.db.Query(ctx, `
		SELECT id, customer_name, customer_phone, total_amount, submitted_at, status
		FROM submissions
		WHERE submitted_at >= $1 AND submitted_at < $2
		ORDER BY submitted_at DESC, id DESC
	`, monthStart, monthEnd)
	if err != nil {
		return nil, fmt.Errorf("list submissions by month: %w", err)
	}
	defer rows.Close()

	var summaries []Summary
	for rows.Next() {
		var summary Summary
		if err := rows.Scan(
			&summary.ID,
			&summary.CustomerName,
			&summary.CustomerPhone,
			&summary.TotalAmount,
			&summary.SubmittedAt,
			&summary.Status,
		); err != nil {
			return nil, fmt.Errorf("scan submission summary: %w", err)
		}
		summaries = append(summaries, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate submission summaries: %w", err)
	}
	return summaries, nil
}

type Detail struct {
	ID           int64
	CustomerName string
	PostalCode   string
	Address      string
	Note         string
	SubmittedAt  time.Time
	Items        []pricing.Item
}

func (repository *Repository) FindDetailsByIDs(ctx context.Context, ids []int64) ([]Detail, error) {
	rows, err := repository.db.Query(ctx, `
		SELECT
			s.id,
			s.customer_name,
			s.postal_code,
			s.address,
			s.note,
			s.submitted_at,
			i.product_id,
			i.product_name_snapshot,
			i.unit_price_snapshot,
			i.quantity,
			i.total_amount_snapshot
		FROM submissions s
		JOIN submission_items i ON i.submission_id = s.id
		WHERE s.id = ANY($1)
		ORDER BY s.id, i.id
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("find submission details: %w", err)
	}
	defer rows.Close()

	var details []Detail
	for rows.Next() {
		var (
			submissionID int64
			item         pricing.Item
		)
		var customerName, postalCode, address, note string
		var submittedAt time.Time
		if err := rows.Scan(
			&submissionID,
			&customerName,
			&postalCode,
			&address,
			&note,
			&submittedAt,
			&item.ProductID,
			&item.Name,
			&item.UnitPrice,
			&item.Quantity,
			&item.Amount,
		); err != nil {
			return nil, fmt.Errorf("scan submission detail: %w", err)
		}

		if len(details) == 0 || details[len(details)-1].ID != submissionID {
			details = append(details, Detail{
				ID:           submissionID,
				CustomerName: customerName,
				PostalCode:   postalCode,
				Address:      address,
				Note:         note,
				SubmittedAt:  submittedAt,
			})
		}
		details[len(details)-1].Items = append(details[len(details)-1].Items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate submission details: %w", err)
	}
	return details, nil
}
