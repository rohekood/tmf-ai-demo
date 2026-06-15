-- Enforce that an email address belongs to at most one party.
-- A partial, case-insensitive unique index over email contact mediums:
--   * scoped to medium_type = 'email' so phone/address rows are unaffected;
--   * LOWER(value) so casing differences cannot create duplicates;
--   * NULL/blank values are excluded.
CREATE UNIQUE INDEX IF NOT EXISTS uq_party_contact_mediums_email
    ON party_contact_mediums (LOWER(value))
    WHERE medium_type = 'email' AND value IS NOT NULL AND value <> '';
