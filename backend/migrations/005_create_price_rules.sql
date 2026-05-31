CREATE TABLE product_price_rules (
    id BIGSERIAL PRIMARY KEY,
    product_id TEXT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    rule_type TEXT NOT NULL CHECK (rule_type IN ('fixed_total', 'tier_unit')),
    min_quantity INTEGER NOT NULL CHECK (min_quantity >= 1),
    max_quantity INTEGER CHECK (max_quantity IS NULL OR max_quantity >= min_quantity),
    unit_price INTEGER CHECK (unit_price IS NULL OR unit_price >= 0),
    total_price INTEGER CHECK (total_price IS NULL OR total_price >= 0),
    priority INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (
        (rule_type = 'fixed_total' AND total_price IS NOT NULL AND unit_price IS NULL)
        OR
        (rule_type = 'tier_unit' AND unit_price IS NOT NULL AND total_price IS NULL)
    )
);

CREATE INDEX product_price_rules_product_id_idx
    ON product_price_rules (product_id, is_active, priority DESC);
