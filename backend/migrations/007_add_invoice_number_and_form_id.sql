ALTER TABLE submissions ADD COLUMN form_id BIGINT REFERENCES forms(id);
ALTER TABLE submissions ADD COLUMN invoice_number TEXT;

UPDATE submissions
SET form_id = (SELECT id FROM forms WHERE public_slug = 'default')
WHERE form_id IS NULL;

UPDATE submissions
SET invoice_number = 'INV-' || to_char(submitted_at AT TIME ZONE 'Asia/Tokyo', 'YYYYMM') || '-' || lpad(id::text, 6, '0')
WHERE invoice_number IS NULL;

ALTER TABLE submissions ALTER COLUMN form_id SET NOT NULL;
ALTER TABLE submissions ALTER COLUMN invoice_number SET NOT NULL;
ALTER TABLE submissions ADD CONSTRAINT submissions_invoice_number_key UNIQUE (invoice_number);
