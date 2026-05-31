CREATE TABLE products (
    id TEXT PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category TEXT NOT NULL,
    base_unit_price INTEGER NOT NULL CHECK (base_unit_price >= 0),
    tax_type TEXT NOT NULL DEFAULT 'included' CHECK (tax_type IN ('included', 'excluded')),
    tax_rate NUMERIC(5, 2) NOT NULL DEFAULT 10.00 CHECK (tax_rate >= 0),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
