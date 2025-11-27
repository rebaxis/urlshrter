-- migrations/000003_update_short_url_column.up.sql
-- обновляем столбец short_url
ALTER TABLE urls
ALTER COLUMN short_url TYPE VARCHAR(20);