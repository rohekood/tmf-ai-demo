-- Audit triggers for POCV
-- Requires the audit schema to exist (we will assume it's created if sharing DB or create it)
-- Since each service is independent, we create schema again.

CREATE SCHEMA IF NOT EXISTS audit;
-- ... (Copying standard audit function) ...
-- For brevity, I'll assume the same function definition as Shopping Cart
-- In a real repo, we might share the SQL file or use a common migration.

CREATE TABLE IF NOT EXISTS audit.logged_actions (
    schema_name text NOT NULL,
    table_name text NOT NULL,
    user_name text,
    action_tstamp_clk timestamp with time zone NOT NULL DEFAULT clock_timestamp(),
    transaction_id bigint,
    action TEXT NOT NULL CHECK (action IN ('I','D','U','T')),
    original_data jsonb,
    new_data jsonb,
    query text
);

CREATE OR REPLACE FUNCTION audit.if_modified_func() RETURNS trigger AS $body$
DECLARE
    audit_row audit.logged_actions;
    v_user text;
BEGIN
    IF TG_WHEN <> 'AFTER' THEN
        RAISE EXCEPTION 'audit.if_modified_func() must be fired as an AFTER trigger';
    END IF;

    v_user := current_setting('app.current_user', true);
    IF v_user IS NULL OR v_user = '' THEN
        v_user := session_user;
    END IF;

    audit_row = ROW(
        TG_TABLE_SCHEMA::text,
        TG_TABLE_NAME::text,
        v_user,
        clock_timestamp(),
        txid_current(),
        substring(TG_OP,1,1),
        NULL, NULL,
        current_query()
    );

    IF (TG_OP = 'UPDATE') THEN
        audit_row.original_data = row_to_json(OLD);
        audit_row.new_data = row_to_json(NEW);
    ELSIF (TG_OP = 'DELETE') THEN
        audit_row.original_data = row_to_json(OLD);
    ELSIF (TG_OP = 'INSERT') THEN
        audit_row.new_data = row_to_json(NEW);
    END IF;

    INSERT INTO audit.logged_actions VALUES (audit_row.*);
    RETURN NULL;
END;
$body$
LANGUAGE plpgsql
SECURITY DEFINER;

CREATE TRIGGER audit_trigger_saga BEFORE INSERT OR UPDATE OR DELETE ON saga_instances FOR EACH ROW EXECUTE PROCEDURE audit.if_modified_func();
