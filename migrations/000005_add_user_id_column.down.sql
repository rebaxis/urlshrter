-- migrations/000005_add_user_id_column.down.sql
-- удаляем столбец user_id
ALTER TABLE urls DROP COLUMN IF EXISTS user_id;
