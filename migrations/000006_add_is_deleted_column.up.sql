-- migrations/000006_add_is_deleted_column.up.sql
-- добавляем столбец is_deleted
ALTER TABLE urls ADD COLUMN is_deleted BOOLEAN DEFAULT FALSE NOT NULL;
