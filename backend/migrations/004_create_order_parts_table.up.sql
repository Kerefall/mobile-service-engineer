-- Создаем таблицу списанных запчастей (связь many-to-many)
CREATE TABLE IF NOT EXISTS order_parts (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    part_id BIGINT NOT NULL REFERENCES parts(id) ON DELETE RESTRICT,
    quantity INT NOT NULL CHECK (quantity > 0),
    price_at_moment DECIMAL(10, 2), -- цена на момент списания (для истории)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Уникальная пара (заказ + запчасть) чтобы не дублировать
    UNIQUE(order_id, part_id)
);

-- Индекс для быстрой загрузки запчастей по заказу
CREATE INDEX idx_order_parts_order_id ON order_parts(order_id);
CREATE INDEX idx_order_parts_part_id ON order_parts(part_id);