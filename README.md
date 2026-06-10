# Application Invoice Form

A web application for accepting product or service orders and generating Excel invoices from submitted order details.

Applicants use a public form to enter products, quantities, and applicant information. Administrators can manage products, forms, and pricing rules, then export Excel invoices from the submission list.

## Features

### For Applicants

- Open a public order form
- Enter product quantities
- Enter name, address, phone number, and other applicant information
- Review entered details and total amount before submitting
- Submit an order

### For Administrators

- Create and edit products
- Create and edit public forms
- Select which products appear on each form
- Create and edit quantity-based pricing rules
- View submissions by month
- View submission details
- Download selected invoices as a ZIP file containing Excel files

## Order Form Flow

```text
Select products
↓
Enter applicant information
↓
Review entered details
↓
Submit order
```

Applicant information fields:

```text
Name (required)
Phonetic name
Postal code (required)
Address (required)
Phone number (required)
Email address
Notes
```

When product quantities are changed, subtotals and the total amount are updated on the screen.

## Admin Features

### Product Management

Manage product name, description, category, tax-included unit price, display order, and active/inactive status.

### Form Management

Manage public form title, description, public slug, displayed products, and active/inactive status.

The public slug is used in the form URL.

```text
/forms/default
/forms/example
```

### Pricing Rule Management

Set pricing rules for each product.

```text
Fixed total by quantity:
  Example: 2 items total 1,900 yen

Unit price by quantity:
  Example: 900 yen per item for 3 or more items
```

If no pricing rule applies to a quantity, the product's regular unit price is used.

### Submission List

View submissions by selecting a month.

The list shows invoice number, form name, submission date and time, name, phone number, total amount, and status.

Selected submissions can be downloaded together as a ZIP file containing Excel invoices.

## Tech Stack

```text
Frontend: React / TypeScript / Vite
Backend: Go
Database: PostgreSQL
Excel: Excelize
Local DB: Docker Compose
```

## Directory Structure

```text
frontend/                React frontend
backend/                 Go backend
backend/migrations/      Database migration SQL files
backend/templates/       Excel invoice template
compose.yaml             PostgreSQL for local development
```

## Local Setup

### 1. Start the database

```bash
docker compose up -d
```

PostgreSQL starts, and the SQL files in `backend/migrations/` are executed during initialization.

### 2. Start the backend

```bash
cd backend
DATABASE_URL='postgres://form_invoice_generator:local_development_only@127.0.0.1:5432/form_invoice_generator?sslmode=disable' go run ./cmd/api
```

Check the backend:

```bash
curl http://127.0.0.1:8080/health
```

If `ok` is returned, the backend is running.

### 3. Start the frontend

Run this in another terminal.

```bash
cd frontend
npm install
npm run dev
```

Open the displayed URL in a browser.

```text
http://127.0.0.1:5173/
```

## Local Pages

```text
Public form:
http://127.0.0.1:5173/

Admin home:
http://127.0.0.1:5173/admin

Product management:
http://127.0.0.1:5173/admin/products

Form management:
http://127.0.0.1:5173/admin/forms

Pricing rule management:
http://127.0.0.1:5173/admin/price-rules

Submission list:
http://127.0.0.1:5173/admin/submissions
```

## Notes

- This repository does not use a `.env` file.
- Do not commit database connection strings or passwords to Git.
