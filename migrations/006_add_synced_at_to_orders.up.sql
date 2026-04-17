-- Добавляем поле synced_at для отслеживания синхронизации
ALTER TABLE orders ADD COLUMN IF NOT EXISTS synced_at TIMESTAMP;

-- Добавляем поле sync_attempts для подсчета попыток синхронизации
ALTER TABLE orders ADD COLUMN IF NOT EXISTS sync_attempts INT DEFAULT 0;

-- Индекс для быстрого поиска несинхронизированных заказов
CREATE INDEX IF NOT EXISTS idx_orders_synced_at ON orders(synced_at) WHERE synced_at IS NULL;