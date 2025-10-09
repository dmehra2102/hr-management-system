-- Rollback: remove employees-related database objects

-- 1. Drop foreign key constraint from departments
ALTER TABLE departments DROP CONSTRAINT IF EXISTS fk_departments_manager;

-- 2. Drop trigger from employees table
DROP TRIGGER IF EXISTS update_employees_updated_at ON employees;

-- 3. Drop indexes explicitly (optional but good practice)
DROP INDEX IF EXISTS idx_employees_employee_id;
DROP INDEX IF EXISTS idx_employees_email;
DROP INDEX IF EXISTS idx_employees_department_id;
DROP INDEX IF EXISTS idx_employees_status;
DROP INDEX IF EXISTS idx_employees_hire_date;
DROP INDEX IF EXISTS idx_employees_name;

-- 4. Finally, drop the employees table
DROP TABLE IF EXISTS employees;
