-- migrations/000002_add_uuid_column.down.sql
-- удаляем столбец uuid
ALTER TABLE urls DROP COLUMN IF EXISTS uuid;
