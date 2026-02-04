-- Employee Profile schema
CREATE TABLE IF NOT EXISTS employees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(50) NOT NULL,
    department VARCHAR(100) NOT NULL,
    position VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'Active',
    employment_type VARCHAR(50) NOT NULL DEFAULT 'FullTime',
    join_date TIMESTAMP WITH TIME ZONE NOT NULL,
    manager_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_employees_status ON employees(status);
CREATE INDEX IF NOT EXISTS idx_employees_department ON employees(department);
CREATE INDEX IF NOT EXISTS idx_employees_manager_id ON employees(manager_id);

-- Comments for documentation
COMMENT ON COLUMN employees.status IS 'Active, OnLeave, Resigned, Terminated';
COMMENT ON COLUMN employees.employment_type IS 'FullTime, PartTime, Contract, Intern';
