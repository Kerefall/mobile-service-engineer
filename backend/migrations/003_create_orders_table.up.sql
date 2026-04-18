-- Создаем таблицу заказ-нарядов
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    address VARCHAR(500) NOT NULL,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    scheduled_date TIMESTAMP NOT NULL,
    status VARCHAR(50) DEFAULT 'new', -- new, in_progress, on_the_way, completed, syncing
    engineer_id BIGINT NOT NULL REFERENCES engineers(id) ON DELETE RESTRICT,
    
    -- Фото
    photo_before_path VARCHAR(500),
    photo_after_path VARCHAR(500),
    
    -- Подпись и акт
    signature_path VARCHAR(500),
    pdf_path VARCHAR(500),
    
    -- Временные метки
    arrival_time TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Синхронизация
    synced_at TIMESTAMP,
    sync_attempts INT DEFAULT 0
);

-- Индексы для быстрых запросов
CREATE INDEX idx_orders_engineer_id ON orders(engineer_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_scheduled_date ON orders(scheduled_date);
CREATE INDEX idx_orders_engineer_status ON orders(engineer_id, status);