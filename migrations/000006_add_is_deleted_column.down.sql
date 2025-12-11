-- migrations/000006_add_is_deleted_column.down.sql
-- удаляем столбец is_deleted
ALTER TABLE urls DROP COLUMN IF EXISTS is_deleted;
