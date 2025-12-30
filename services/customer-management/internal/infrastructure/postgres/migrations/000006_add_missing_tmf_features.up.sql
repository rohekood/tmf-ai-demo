CREATE TABLE IF NOT EXISTS related_parties (
    id VARCHAR(255) PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    related_party_id VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    valid_for_start TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    valid_for_end TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_related_parties_customer_id ON related_parties(customer_id);

CREATE TABLE IF NOT EXISTS payment_methods (
    id VARCHAR(255) PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    token VARCHAR(255) NOT NULL,
    details JSONB,
    is_default BOOLEAN DEFAULT FALSE,
    valid_for_start TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    valid_for_end TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_payment_methods_customer_id ON payment_methods(customer_id);

CREATE TABLE IF NOT EXISTS market_segments (
    id VARCHAR(255) PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    category VARCHAR(100)
);

CREATE INDEX IF NOT EXISTS idx_market_segments_customer_id ON market_segments(customer_id);

CREATE TABLE IF NOT EXISTS customer_interactions (
    id VARCHAR(255) PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    interaction_date TIMESTAMP WITH TIME ZONE NOT NULL,
    channel VARCHAR(50) NOT NULL,
    type VARCHAR(50) NOT NULL,
    description TEXT,
    agent_id VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_customer_interactions_customer_id ON customer_interactions(customer_id);

CREATE TABLE IF NOT EXISTS applied_billing_rates (
    id VARCHAR(255) PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    product_ref VARCHAR(100) NOT NULL,
    rate_type VARCHAR(50) NOT NULL,
    value DECIMAL(10, 2) NOT NULL,
    valid_for_start TIMESTAMP WITH TIME ZONE NOT NULL,
    valid_for_end TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_applied_billing_rates_customer_id ON applied_billing_rates(customer_id);

ALTER TABLE customer_accounts 
ADD COLUMN IF NOT EXISTS bill_format VARCHAR(50) DEFAULT 'PDF',
ADD COLUMN IF NOT EXISTS billing_cycle VARCHAR(50) DEFAULT 'Monthly';
