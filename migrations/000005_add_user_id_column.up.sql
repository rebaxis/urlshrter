-- migrations/000005_add_user_id_column.up.sql
-- добавляем столбец user_id
ALTER TABLE urls ADD COLUMN user_id VARCHAR(20) DEFAULT ''; 
ALTER TABLE urls ALTER COLUMN user_id SET NOT NULL;
