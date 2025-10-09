-- Drop employees table
ALTER TABLE departments DROP CONSTRAINT IF EXISTS fk_departments_manager;
DROP TRIGGER IF EXISTS update_employees_updated_at ON expmployees;
DROP TABLE IF EXISTS employees;