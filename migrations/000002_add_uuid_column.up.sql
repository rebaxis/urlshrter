-- migrations/000002_add_uuid_column.up.sql
-- добавляем столбец uuid
ALTER TABLE urls ADD COLUMN uuid VARCHAR(50); 

ALTER TABLE urls ALTER COLUMN uuid SET NOT NULL;
