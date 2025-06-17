BEGIN;

CREATE TYPE customer_status AS ENUM ('ACTIVE', 'SUSPENDED');

ALTER TABLE customers
ADD COLUMN status customer_status NOT NULL;


COMMIT;