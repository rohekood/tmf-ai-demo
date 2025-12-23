-- Contact Mediums for Parties
CREATE TABLE IF NOT EXISTS party_contact_mediums (
    id VARCHAR(255) PRIMARY KEY,
    party_id VARCHAR(255) NOT NULL REFERENCES parties(id) ON DELETE CASCADE,
    medium_type VARCHAR(50),
    preferred BOOLEAN,
    value TEXT,
    street1 TEXT,
    street2 TEXT,
    city TEXT,
    state_or_province TEXT,
    postcode TEXT,
    country TEXT,
    valid_for_start TIMESTAMP WITH TIME ZONE,
    valid_for_end TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Identifications (Passport, National ID, etc.)
CREATE TABLE IF NOT EXISTS identifications (
    id VARCHAR(255) PRIMARY KEY,
    party_id VARCHAR(255) NOT NULL REFERENCES parties(id) ON DELETE CASCADE,
    identification_type VARCHAR(255) NOT NULL,
    identification_id VARCHAR(255) NOT NULL,
    issuing_authority VARCHAR(255),
    issuing_date TIMESTAMP WITH TIME ZONE,
    valid_for_start TIMESTAMP WITH TIME ZONE,
    valid_for_end TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Related Parties (Relationships between individuals/organizations)
CREATE TABLE IF NOT EXISTS related_parties (
    id VARCHAR(255) PRIMARY KEY,
    party_id VARCHAR(255) NOT NULL REFERENCES parties(id) ON DELETE CASCADE,
    related_party_id VARCHAR(255) NOT NULL, -- The target party
    related_party_name VARCHAR(255),
    role VARCHAR(255),
    valid_for_start TIMESTAMP WITH TIME ZONE,
    valid_for_end TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_party_contact_mediums_party_id ON party_contact_mediums(party_id);
CREATE INDEX IF NOT EXISTS idx_identifications_party_id ON identifications(party_id);
CREATE INDEX IF NOT EXISTS idx_related_parties_party_id ON related_parties(party_id);
