-- Drop departments table

DROP TRIGGER IF EXISTS update_departments_updated_at ON departments;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS departments;