ALTER TABLE related_parties 
DROP COLUMN IF EXISTS permissions,
DROP COLUMN IF EXISTS role;

DROP TABLE IF EXISTS party_attachments;
DROP TABLE IF EXISTS attachment_contents;
DROP TABLE IF EXISTS party_tax_exemptions;
DROP TABLE IF EXISTS external_references;
