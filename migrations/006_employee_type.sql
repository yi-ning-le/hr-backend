-- Add employee_type column to differentiate HR from regular employees
ALTER TABLE employees 
ADD COLUMN IF NOT EXISTS employee_type VARCHAR(20) NOT NULL DEFAULT 'EMPLOYEE';

COMMENT ON COLUMN employees.employee_type IS 'HR, EMPLOYEE';
