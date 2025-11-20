-- migrations/000003_update_short_url_column.down.sql
ALTER TABLE urls
ALTER COLUMN short_url TYPE VARCHAR(255);
