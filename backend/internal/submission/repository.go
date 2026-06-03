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
	FormID        int64
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

type Created struct {
	ID            int64
	InvoiceNumber string
	SubmittedAt   time.Time
}

func (repository *Repository) Create(ctx context.Context, submission Submission) (Created, error) {
	transaction, err := repository.db.Begin(ctx)
	if err != nil {
		return Created{}, fmt.Errorf("begin submission transaction: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	var created Created
	err = transaction.QueryRow(ctx, `
		SELECT nextval(pg_get_serial_sequence('submissions', 'id')), CURRENT_TIMESTAMP
	`).Scan(&created.ID, &created.SubmittedAt)
	if err != nil {
		return Created{}, fmt.Errorf("reserve submission id: %w", err)
	}
	created.InvoiceNumber = FormatInvoiceNumber(created.ID, created.SubmittedAt)

	_, err = transaction.Exec(ctx, `
		INSERT INTO submissions (
			id,
			form_id,
			invoice_number,
			customer_name,
			customer_kana,
			customer_email,
			customer_phone,
			postal_code,
			address,
			note,
			total_amount,
			submitted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`,
		created.ID,
		submission.FormID,
		created.InvoiceNumber,
		submission.CustomerName,
		submission.CustomerKana,
		submission.CustomerEmail,
		submission.CustomerPhone,
		submission.PostalCode,
		submission.Address,
		submission.Note,
		submission.TotalAmount,
		created.SubmittedAt,
	)
	if err != nil {
		return Created{}, fmt.Errorf("insert submission: %w", err)
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
			created.ID,
			item.ProductID,
			item.Name,
			item.Quantity,
			item.UnitPrice,
			item.Amount,
		)
		if err != nil {
			return Created{}, fmt.Errorf("insert submission item: %w", err)
		}
	}

	if err := transaction.Commit(ctx); err != nil {
		return Created{}, fmt.Errorf("commit submission transaction: %w", err)
	}
	return created, nil
}

func FormatInvoiceNumber(id int64, submittedAt time.Time) string {
	jst := time.FixedZone("JST", 9*60*60)
	return fmt.Sprintf("INV-%s-%06d", submittedAt.In(jst).Format("200601"), id)
}

type Summary struct {
	ID            int64     `json:"id"`
	InvoiceNumber string    `json:"invoiceNumber"`
	FormTitle     string    `json:"formTitle"`
	FormSlug      string    `json:"formSlug"`
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
		SELECT s.id, s.invoice_number, f.title, f.public_slug, s.customer_name, s.customer_phone, s.total_amount, s.submitted_at, s.status
		FROM submissions s
		JOIN forms f ON f.id = s.form_id
		WHERE s.submitted_at >= $1 AND s.submitted_at < $2
		ORDER BY s.submitted_at DESC, s.id DESC
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
			&summary.InvoiceNumber,
			&summary.FormTitle,
			&summary.FormSlug,
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
	ID            int64          `json:"id"`
	InvoiceNumber string         `json:"invoiceNumber"`
	CustomerName  string         `json:"customerName"`
	PostalCode    string         `json:"postalCode"`
	Address       string         `json:"address"`
	Note          string         `json:"note"`
	SubmittedAt   time.Time      `json:"submittedAt"`
	Items         []pricing.Item `json:"items"`
}

func (repository *Repository) FindDetailByID(ctx context.Context, id int64) (Detail, error) {
	details, err := repository.FindDetailsByIDs(ctx, []int64{id})
	if err != nil {
		return Detail{}, err
	}
	if len(details) == 0 {
		return Detail{}, fmt.Errorf("submission not found: %d", id)
	}
	return details[0], nil
}

func (repository *Repository) FindDetailsByIDs(ctx context.Context, ids []int64) ([]Detail, error) {
	rows, err := repository.db.Query(ctx, `
		SELECT
			s.id,
			s.invoice_number,
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
		var detailsInvoiceNumber string
		var submittedAt time.Time
		if err := rows.Scan(
			&submissionID,
			&detailsInvoiceNumber,
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
				ID:            submissionID,
				InvoiceNumber: detailsInvoiceNumber,
				CustomerName:  customerName,
				PostalCode:    postalCode,
				Address:       address,
				Note:          note,
				SubmittedAt:   submittedAt,
			})
		}
		details[len(details)-1].Items = append(details[len(details)-1].Items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate submission details: %w", err)
	}
	return details, nil
}
