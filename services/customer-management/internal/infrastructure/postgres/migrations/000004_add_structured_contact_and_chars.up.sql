-- Add structured fields to contact_mediums
ALTER TABLE contact_mediums ADD COLUMN IF NOT EXISTS street1 TEXT;
ALTER TABLE contact_mediums ADD COLUMN IF NOT EXISTS street2 TEXT;
ALTER TABLE contact_mediums ADD COLUMN IF NOT EXISTS city TEXT;
ALTER TABLE contact_mediums ADD COLUMN IF NOT EXISTS state_or_province TEXT;
ALTER TABLE contact_mediums ADD COLUMN IF NOT EXISTS postcode TEXT;
ALTER TABLE contact_mediums ADD COLUMN IF NOT EXISTS country TEXT;

-- Create customer_characteristics table for dynamic extension
CREATE TABLE IF NOT EXISTS customer_characteristics (
    id VARCHAR(255) PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    value TEXT,
    value_type VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_customer_characteristics_customer_id ON customer_characteristics(customer_id);
