CREATE TABLE IF NOT EXISTS tax_exemptions (
    id UUID PRIMARY KEY,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    certificate_number VARCHAR(255),
    issuing_jurisdiction VARCHAR(255),
    valid_for_start TIMESTAMPTZ,
    valid_for_end TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_tax_exemptions_customer_id ON tax_exemptions(customer_id);

CREATE TRIGGER update_tax_exemptions_modtime
    BEFORE UPDATE ON tax_exemptions
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

SELECT audit.audit_table('tax_exemptions');
