-- Qualification Sessions Table
-- Stores persistent qualification sessions with customer-specific prices

CREATE TABLE IF NOT EXISTS qualification_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL,
    address JSONB NOT NULL,
    qualified_offers JSONB NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL DEFAULT (NOW() + INTERVAL '24 hours')
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_qualification_sessions_customer_id ON qualification_sessions(customer_id);
CREATE INDEX IF NOT EXISTS idx_qualification_sessions_expires_at ON qualification_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_qualification_sessions_created_at ON qualification_sessions(created_at);

-- Comments for documentation
COMMENT ON TABLE qualification_sessions IS 'TMF679 Product Offering Qualification sessions with customer-specific pricing';
COMMENT ON COLUMN qualification_sessions.id IS 'Unique session identifier';
COMMENT ON COLUMN qualification_sessions.customer_id IS 'Reference to customer who requested qualification';
COMMENT ON COLUMN qualification_sessions.address IS 'Address that was qualified (JSONB for flexibility)';
COMMENT ON COLUMN qualification_sessions.qualified_offers IS 'Array of qualified offerings with customer-specific prices';
COMMENT ON COLUMN qualification_sessions.status IS 'QUALIFIED, UNQUALIFIED, or EXPIRED';
COMMENT ON COLUMN qualification_sessions.expires_at IS 'Session expiry time (default 24 hours from creation)';
