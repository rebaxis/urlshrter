-- migrations/000001_create_urls_table.up.sql
-- Создание таблицы коротких URL
CREATE TABLE urls (
    short_url VARCHAR(255) NOT NULL,
    original_url VARCHAR(255) NOT NULL
);
