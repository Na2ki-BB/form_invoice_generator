CREATE TABLE forms (
    id BIGSERIAL PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    public_slug TEXT NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE form_products (
    id BIGSERIAL PRIMARY KEY,
    form_id BIGINT NOT NULL REFERENCES forms(id) ON DELETE CASCADE,
    product_id TEXT NOT NULL REFERENCES products(id),
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_required BOOLEAN NOT NULL DEFAULT FALSE,
    min_quantity INTEGER NOT NULL DEFAULT 0 CHECK (min_quantity >= 0),
    max_quantity INTEGER NOT NULL DEFAULT 10 CHECK (max_quantity >= min_quantity),
    UNIQUE (form_id, product_id)
);
