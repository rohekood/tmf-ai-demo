-- Anonymous availability checks have no customer attached, so customer_id must
-- be optional. Customer-specific pricing still populates it when present.
ALTER TABLE qualification_sessions ALTER COLUMN customer_id DROP NOT NULL;
