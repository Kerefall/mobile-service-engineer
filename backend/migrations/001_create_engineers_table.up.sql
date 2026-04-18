-- Создаем таблицу инженеров
CREATE TABLE IF NOT EXISTS engineers (
    id BIGSERIAL PRIMARY KEY,
    login VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL, -- хеш пароля (bcrypt)
    fcm_token VARCHAR(255), -- токен для push-уведомлений
    full_name VARCHAR(255),
    phone VARCHAR(20),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индекс для быстрого поиска по логину
CREATE INDEX idx_engineers_login ON engineers(login);