-- migrations/000004_create_index_for_original_url.up.sql
-- Удаляем индекс для original_url
DROP INDEX IF EXISTS original_url_idx;
ALTER TABLE urls DROP CONSTRAINT uniq_original_url;
