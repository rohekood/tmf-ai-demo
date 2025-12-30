CREATE TABLE IF NOT EXISTS external_references (
    id UUID PRIMARY KEY,
    party_id VARCHAR(255) NOT NULL REFERENCES parties(id) ON DELETE CASCADE,
    external_system_id VARCHAR(100) NOT NULL,
    external_reference_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ext_ref ON external_references (external_system_id, external_reference_id);

CREATE TABLE IF NOT EXISTS party_tax_exemptions (
    id UUID PRIMARY KEY,
    party_id VARCHAR(255) NOT NULL REFERENCES parties(id) ON DELETE CASCADE,
    certificate_number VARCHAR(100),
    issuing_jurisdiction VARCHAR(100),
    valid_for_start TIMESTAMP WITH TIME ZONE,
    valid_for_end TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Content table (Blob storage)
CREATE TABLE IF NOT EXISTS attachment_contents (
    id UUID PRIMARY KEY,
    data BYTEA,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Metadata table (Polymorphic reference)
CREATE TABLE IF NOT EXISTS party_attachments (
    id UUID PRIMARY KEY,
    owner_id VARCHAR(255) NOT NULL REFERENCES parties(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    attachment_type VARCHAR(50),
    ref_type VARCHAR(50) NOT NULL, -- 'Internal' or 'S3'
    ref_id VARCHAR(255) NOT NULL,  -- UUID or URL/Key
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

ALTER TABLE related_parties 
ADD COLUMN IF NOT EXISTS role VARCHAR(50),
ADD COLUMN IF NOT EXISTS permissions JSONB;
