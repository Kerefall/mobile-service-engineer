DROP INDEX IF EXISTS idx_orders_synced_at;
ALTER TABLE orders DROP COLUMN IF EXISTS synced_at;
ALTER TABLE orders DROP COLUMN IF EXISTS sync_attempts;