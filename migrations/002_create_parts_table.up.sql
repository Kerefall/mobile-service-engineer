-- Создаем справочник запчастей
CREATE TABLE IF NOT EXISTS parts (
    id BIGSERIAL PRIMARY KEY,
    article VARCHAR(100) UNIQUE NOT NULL, -- артикул/код запчасти
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) DEFAULT 0,
    quantity_in_stock INT DEFAULT 0, -- остаток на складе
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индекс для поиска по артикулу
CREATE INDEX idx_parts_article ON parts(article);
CREATE INDEX idx_parts_name ON parts(name);