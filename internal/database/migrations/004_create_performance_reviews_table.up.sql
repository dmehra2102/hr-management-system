CREATE TABLE IF NOT EXISTS performance_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    reviewer_id UUID NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    review_period VARCHAR(20) NOT NULL CHECK (review_period IN ('QUARTERLY','HALF_YEALY','ANNUAL', 'PROBATION')),
    review_date DATE NOT NULL,
    status VARCHAR(20) DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'SUBMITTED', 'COMPLETED', 'ARCHIVED')),
    overall_rating DECIMAL(3, 2) CHECK (overall_rating >= 0 AND overall_rating <= 5),
    overall_comments TEXT,
    submitted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS performance_goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id UUID NOT NULL REFERENCES performance_reviews(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    target_value DECIMAL(10,2),
    achieved_value DECIMAL(10,2),
    uint VARCHAR(50),
    status VARCHAR(20) DEFAULT 'NOT STARTED' CHECK (status IN ('NOT_STARTED','IN_PROGRESS','COMPLETED', 'EXCEEDED', 'NOT_ACHIEVED')),
    weight DECIMAL(5, 2) DEFAULT 1.00 CHECK (weight >= 0 AND weight <= 1),
    comments TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS performance_competencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id UUID NOT NULL REFERENCES performance_reviews(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    rating DECIMAL(3, 2) CHECK (rating >= 0 AND rating <= 5),
    max_rating DECIMAL(3, 2) DEFAULT 5.00,
    weight DECIMAL(5, 2) DEFAULT 1.00 CHECK (weight >= 0 AND weight <= 1),
    comments TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_performance_reviews_employee_id ON performance_reviews(employee_id);
CREATE INDEX IF NOT EXISTS idx_performance_reviews_reviewer_id ON performance_reviews(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_performance_reviews_status ON performance_reviews(status);
CREATE INDEX IF NOT EXISTS idx_performance_reviews_review_date ON performance_reviews(review_date);
CREATE INDEX IF NOT EXISTS idx_performance_reviews_period ON performance_reviews(review_period);

CREATE INDEX IF NOT EXISTS idx_performance_goals_review_id ON performance_goals(review_id);
CREATE INDEX IF NOT EXISTS idx_performance_goals_status ON performance_goals(status);

CREATE INDEX IF NOT EXISTS idx_performance_competencies_review_id ON performance_competencies(review_id);

-- Triggers
CREATE TRIGGER update_performance_reviews_updated_at
    BEFORE UPDATE ON performance_reviews
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_performance_goals_updated_at
    BEFORE UPDATE ON performance_goals
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_performance_competencies_updated_at
    BEFORE UPDATE ON performance_competencies
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();