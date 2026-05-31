INSERT INTO product_price_rules (
    product_id,
    rule_type,
    min_quantity,
    max_quantity,
    total_price,
    priority
)
SELECT 'ofuda', 'fixed_total', 2, 2, 1900, 10
WHERE NOT EXISTS (
    SELECT 1
    FROM product_price_rules
    WHERE product_id = 'ofuda'
      AND rule_type = 'fixed_total'
      AND min_quantity = 2
      AND max_quantity = 2
      AND total_price = 1900
);
