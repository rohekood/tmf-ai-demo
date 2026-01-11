ALTER TABLE outbox_events ALTER COLUMN id SET DEFAULT gen_random_uuid();
