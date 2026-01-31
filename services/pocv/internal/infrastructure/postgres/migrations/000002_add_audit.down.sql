DROP TRIGGER IF EXISTS audit_trigger_saga ON saga_instances;
-- We don't drop the function/schema as other services might use it if shared DB, 
-- but strictly this service owns its schema.
