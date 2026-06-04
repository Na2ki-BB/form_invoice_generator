INSERT INTO product_price_rules (
    product_id,
    rule_type,
    min_quantity,
    max_quantity,
    unit_price,
    priority
)
SELECT 'ofuda', 'tier_unit', 2, NULL, 950, 5
WHERE NOT EXISTS (
    SELECT 1
    FROM product_price_rules
    WHERE product_id = 'ofuda'
      AND rule_type = 'tier_unit'
      AND min_quantity = 2
      AND max_quantity IS NULL
      AND unit_price = 950
);
