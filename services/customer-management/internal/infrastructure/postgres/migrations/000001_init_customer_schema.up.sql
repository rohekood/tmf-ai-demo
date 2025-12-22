CREATE TABLE IF NOT EXISTS customers (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255),
    status VARCHAR(50) NOT NULL,
    status_reason TEXT,
    valid_for_start TIMESTAMP WITH TIME ZONE,
    valid_for_end TIMESTAMP WITH TIME ZONE,
    party_id VARCHAR(255) NOT NULL,
    party_type VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS customer_accounts (
    id VARCHAR(255) PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    name VARCHAR(255),
    account_status VARCHAR(50),
    account_type VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS credit_profiles (
    id VARCHAR(255) PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    credit_profile_date TIMESTAMP WITH TIME ZONE,
    credit_risk_score INTEGER,
    credit_score INTEGER,
    valid_for_start TIMESTAMP WITH TIME ZONE,
    valid_for_end TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS contact_mediums (
    id VARCHAR(255) PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    medium_type VARCHAR(50),
    preferred BOOLEAN,
    value TEXT,
    valid_for_start TIMESTAMP WITH TIME ZONE,
    valid_for_end TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_customers_party_id ON customers(party_id);
CREATE INDEX IF NOT EXISTS idx_customer_accounts_customer_id ON customer_accounts(customer_id);
CREATE INDEX IF NOT EXISTS idx_credit_profiles_customer_id ON credit_profiles(customer_id);
CREATE INDEX IF NOT EXISTS idx_contact_mediums_customer_id ON contact_mediums(customer_id);
