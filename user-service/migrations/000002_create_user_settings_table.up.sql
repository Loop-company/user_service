-- Таблица настроек пользователей
CREATE TABLE user_settings
(
    user_id    UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    settings   JSONB NOT NULL           DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индекс для быстрого поиска по JSON полям (GIN индекс)
CREATE INDEX idx_user_settings_gin ON user_settings USING GIN (settings);

-- Триггер для авто-обновления updated_at
CREATE TRIGGER update_user_settings_updated_at
    BEFORE UPDATE
    ON user_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();