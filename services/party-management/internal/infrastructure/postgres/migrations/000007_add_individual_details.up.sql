ALTER TABLE individuals ADD COLUMN IF NOT EXISTS middle_name VARCHAR(255);
ALTER TABLE individuals ADD COLUMN IF NOT EXISTS birth_date VARCHAR(255); -- Keeping as string for simplicity/flexibility as per discussion
ALTER TABLE individuals ADD COLUMN IF NOT EXISTS gender VARCHAR(50);
