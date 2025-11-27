-- migrations/000001_create_urls_table.down.sql
-- Откат создания таблицы коротких URL
DROP TABLE IF EXISTS urls;
