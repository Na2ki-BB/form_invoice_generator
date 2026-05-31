INSERT INTO products (id, owner_id, name, description, category, base_unit_price, sort_order)
VALUES
    ('prayer-a', 1, '祈祷A', '基本の祈祷メニューです。', '祈祷', 5000, 10),
    ('ofuda', 1, '御札', '授与品の御札です。', '授与品', 1000, 20),
    ('omamori', 1, 'お守り', '授与品のお守りです。', '授与品', 800, 30)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    base_unit_price = EXCLUDED.base_unit_price,
    sort_order = EXCLUDED.sort_order,
    updated_at = CURRENT_TIMESTAMP;

INSERT INTO forms (owner_id, title, description, public_slug)
VALUES (1, '申込み請求フォーム', '固定商品を使った動作確認用フォームです。', 'default')
ON CONFLICT (public_slug) DO UPDATE SET
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    updated_at = CURRENT_TIMESTAMP;

INSERT INTO form_products (form_id, product_id, sort_order, min_quantity, max_quantity)
SELECT forms.id, products.id, products.sort_order, 0, 10
FROM forms
JOIN products ON products.id IN ('prayer-a', 'ofuda', 'omamori')
WHERE forms.public_slug = 'default'
ON CONFLICT (form_id, product_id) DO UPDATE SET
    sort_order = EXCLUDED.sort_order,
    min_quantity = EXCLUDED.min_quantity,
    max_quantity = EXCLUDED.max_quantity;
