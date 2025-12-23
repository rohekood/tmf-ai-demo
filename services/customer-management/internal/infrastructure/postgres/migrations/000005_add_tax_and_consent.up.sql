-- Tax Exemptions
CREATE TABLE IF NOT EXISTS tax_exemptions (
    id VARCHAR(255) PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    certificate_number VARCHAR(255),
    issuing_jurisdiction VARCHAR(255),
    valid_for_start TIMESTAMP WITH TIME ZONE,
    valid_for_end TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tax_exemptions_customer_id ON tax_exemptions(customer_id);

-- Privacy Consents
CREATE TABLE IF NOT EXISTS privacy_consents (
    id VARCHAR(255) PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    consent_type VARCHAR(255) NOT NULL, -- e.g. "Marketing", "DataProcessing"
    status VARCHAR(50), -- "Given", "Refused", "Withdrawn"
    valid_for_start TIMESTAMP WITH TIME ZONE,
    valid_for_end TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_privacy_consents_customer_id ON privacy_consents(customer_id);
