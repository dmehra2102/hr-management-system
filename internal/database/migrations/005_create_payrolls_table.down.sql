DROP TRIGGER IF EXISTS track_payroll_changes_trigger ON payroll;
DROP FUNCTION IF EXISTS track_payroll_changes();
DROP TRIGGER IF EXISTS update_payroll_updated_at ON payroll;
DROP TABLE IF EXISTS payroll_history;
DROP TABLE IF EXISTS payroll;