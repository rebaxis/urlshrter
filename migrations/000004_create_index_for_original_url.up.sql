-- migrations/000004_create_index_for_original_url.up.sql
-- Добавляем индекс и отдельно ограничение уникальности для original_url
ALTER TABLE urls ADD CONSTRAINT uniq_original_url UNIQUE (original_url);
CREATE INDEX original_url_idx ON urls (original_url);
