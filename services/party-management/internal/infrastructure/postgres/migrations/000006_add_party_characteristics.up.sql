CREATE TABLE IF NOT EXISTS party_characteristics (
    id VARCHAR(255) PRIMARY KEY,
    party_id VARCHAR(255) NOT NULL REFERENCES parties(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    value_type VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_party_characteristics_party_id ON party_characteristics(party_id);
