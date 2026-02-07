-- Create candidate_statuses table
CREATE TABLE IF NOT EXISTS candidate_statuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    slug VARCHAR(50) NOT NULL UNIQUE, -- Stable identifier for system logic
    type VARCHAR(20) NOT NULL DEFAULT 'custom', -- 'system' or 'custom'
    sort_order INTEGER NOT NULL,
    color VARCHAR(20) NOT NULL, -- Hex color code
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Seed default statuses
INSERT INTO candidate_statuses (name, slug, type, sort_order, color) VALUES
('New', 'new', 'system', 1, '#3b82f6'),        -- Blue
('Screening', 'screening', 'system', 2, '#8b5cf6'), -- Violet
('Interview', 'interview', 'system', 3, '#f59e0b'), -- Amber
('Offer', 'offer', 'system', 4, '#10b981'),      -- Emerald
('Hired', 'hired', 'system', 5, '#15803d'),      -- Green
('Rejected', 'rejected', 'system', 6, '#ef4444') -- Red
ON CONFLICT (slug) DO NOTHING;

-- Index for sorting
CREATE INDEX IF NOT EXISTS idx_candidate_statuses_sort_order ON candidate_statuses(sort_order);
