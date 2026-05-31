CREATE TABLE submissions (
    id BIGSERIAL PRIMARY KEY,
    customer_name TEXT NOT NULL,
    customer_kana TEXT NOT NULL DEFAULT '',
    customer_email TEXT NOT NULL DEFAULT '',
    customer_phone TEXT NOT NULL,
    postal_code TEXT NOT NULL,
    address TEXT NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    total_amount INTEGER NOT NULL CHECK (total_amount >= 0),
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status TEXT NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'checked', 'invoiced', 'cancelled'))
);

CREATE TABLE submission_items (
    id BIGSERIAL PRIMARY KEY,
    submission_id BIGINT NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    product_id TEXT NOT NULL,
    product_name_snapshot TEXT NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price_snapshot INTEGER NOT NULL CHECK (unit_price_snapshot >= 0),
    total_amount_snapshot INTEGER NOT NULL CHECK (total_amount_snapshot >= 0)
);
