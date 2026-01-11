ALTER TABLE outbox_events
  DROP COLUMN aggregate_type,
  DROP COLUMN aggregate_id,
  DROP COLUMN type;

ALTER TABLE outbox_events
  ADD COLUMN routing_key VARCHAR(255) NOT NULL DEFAULT '',
  ADD COLUMN headers JSONB;

ALTER TABLE outbox_events
  ALTER COLUMN routing_key DROP DEFAULT;
