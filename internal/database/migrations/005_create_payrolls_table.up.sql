CREATE TABLE IF NOT EXISTS payroll (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    pay_period_start DATE NOT NULL,
    pay_period_end DATE NOT NULL,
    pay_date DATE NOT NULL,
    
    -- Earnings
    basic_salary DECIMAL(15, 2) NOT NULL DEFAULT 0,
    overtime_hours DECIMAL(8, 2) DEFAULT 0,
    overtime_rate DECIMAL(10, 2) DEFAULT 0,
    overtime_pay DECIMAL(15, 2) DEFAULT 0,
    bonus DECIMAL(15, 2) DEFAULT 0,
    commission DECIMAL(15, 2) DEFAULT 0,
    allowances DECIMAL(15, 2) DEFAULT 0,
    gross_pay DECIMAL(15, 2) GENERATED ALWAYS AS (
        basic_salary + overtime_pay + bonus + commission + allowances
    ) STORED,
    
    -- Deductions
    tax_federal DECIMAL(15, 2) DEFAULT 0,
    tax_state DECIMAL(15, 2) DEFAULT 0,
    tax_social_security DECIMAL(15, 2) DEFAULT 0,
    tax_medicare DECIMAL(15, 2) DEFAULT 0,
    insurance_health DECIMAL(15, 2) DEFAULT 0,
    insurance_dental DECIMAL(15, 2) DEFAULT 0,
    insurance_vision DECIMAL(15, 2) DEFAULT 0,
    retirement_401k DECIMAL(15, 2) DEFAULT 0,
    other_deductions DECIMAL(15, 2) DEFAULT 0,
    total_deductions DECIMAL(15, 2) GENERATED ALWAYS AS (
        tax_federal + tax_state + tax_social_security + tax_medicare + 
        insurance_health + insurance_dental + insurance_vision + 
        retirement_401k + other_deductions
    ) STORED,
    
    -- Net pay
    net_pay DECIMAL(15, 2),
    
    -- Status and metadata
    status VARCHAR(20) DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'PROCESSED', 'PAID', 'CANCELLED')),
    processed_by UUID REFERENCES employees(id) ON DELETE SET NULL,
    processed_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(employee_id, pay_period_start, pay_period_end),

    -- Constraints
    CONSTRAINT valid_pay_period CHECK (pay_period_end >= pay_period_start),
    CONSTRAINT non_negative_salary CHECK (basic_salary >= 0),
    CONSTRAINT non_negative_overtime_hours CHECK (overtime_hours >= 0),
    CONSTRAINT non_negative_overtime_rate CHECK (overtime_rate >= 0)
);

-- Create payroll history for tracking changes
CREATE TABLE IF NOT EXISTS payroll_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payroll_id UUID NOT NULL REFERENCES payroll(id) ON DELETE CASCADE,
    changed_by UUID NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    change_type VARCHAR(20) NOT NULL CHECK (change_type IN ('CREATED', 'UPDATED', 'PROCESSED', 'PAID', 'CANCELLED')),
    old_values JSONB,
    new_values JSONB,
    change_reason TEXT,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_payroll_employee_id ON payroll(employee_id);
CREATE INDEX IF NOT EXISTS idx_payroll_pay_period ON payroll(pay_period_start, pay_period_end);
CREATE INDEX IF NOT EXISTS idx_payroll_pay_date ON payroll(pay_date);
CREATE INDEX IF NOT EXISTS idx_payroll_status ON payroll(status);
CREATE INDEX IF NOT EXISTS idx_payroll_processed_by ON payroll(processed_by);

CREATE INDEX IF NOT EXISTS idx_payroll_history_payroll_id ON payroll_history(payroll_id);
CREATE INDEX IF NOT EXISTS idx_payroll_history_changed_by ON payroll_history(changed_by);
CREATE INDEX IF NOT EXISTS idx_payroll_history_changed_at ON payroll_history(changed_at);

-- Create triggers
CREATE TRIGGER update_payroll_updated_at
    BEFORE UPDATE ON payroll
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create function to track payroll changes
CREATE OR REPLACE FUNCTION track_payroll_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO payroll_history (payroll_id, changed_by, change_type, new_values)
        VALUES (NEW.id, NEW.processed_by, 'CREATED', to_jsonb(NEW));
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO payroll_history (payroll_id, changed_by, change_type, old_values, new_values)
        VALUES (NEW.id, NEW.processed_by, 'UPDATED', to_jsonb(OLD), to_jsonb(NEW));
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ language 'plpgsql';

-- Create trigger to track payroll changes
CREATE TRIGGER track_payroll_changes_trigger
    AFTER INSERT OR UPDATE ON payroll
    FOR EACH ROW
    EXECUTE FUNCTION track_payroll_changes();

-- Create function to calculate net pay
CREATE OR REPLACE FUNCTION calculate_net_pay()
RETURNS TRIGGER AS $$
BEGIN
    NEW.net_pay := (NEW.basic_salary + NEW.overtime_pay + NEW.bonus + NEW.commission + NEW.allowances) - (NEW.tax_federal + NEW.tax_state + NEW.tax_social_security + NEW.tax_medicare + NEW.insurance_health + NEW.insurance_dental + NEW.insurance_vision + NEW.retirement_401k + NEW.other_deductions);
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER calculate_net_pay_before_insert_update
    BEFORE INSERT OR UPDATE ON payroll
    FOR EACH ROW
    EXECUTE FUNCTION calculate_net_pay();