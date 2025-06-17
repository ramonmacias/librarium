BEGIN;

ALTER TABLE customers
DROP COLUMN status;

DROP TYPE customer_status;

COMMIT;