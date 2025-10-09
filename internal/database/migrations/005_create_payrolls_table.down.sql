-- Drop triggers
DROP TRIGGER IF EXISTS track_payroll_changes_trigger ON payroll;
DROP TRIGGER IF EXISTS update_payroll_updated_at ON payroll;
DROP TRIGGER IF EXISTS calculate_net_pay_before_insert_update ON payroll;

-- Drop trigger function
DROP FUNCTION IF EXISTS track_payroll_changes();
DROP FUNCTION IF EXISTS calculate_net_pay();

-- Drop history table first (depends on payroll)
DROP TABLE IF EXISTS payroll_history;

-- Drop main payroll table
DROP TABLE IF EXISTS payroll;