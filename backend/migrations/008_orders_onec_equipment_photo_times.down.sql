DROP INDEX IF EXISTS idx_orders_onec_guid;
ALTER TABLE orders DROP COLUMN IF EXISTS photo_after_at;
ALTER TABLE orders DROP COLUMN IF EXISTS photo_before_at;
ALTER TABLE orders DROP COLUMN IF EXISTS onec_guid;
ALTER TABLE orders DROP COLUMN IF EXISTS equipment;
