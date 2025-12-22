-- Based on https://wiki.postgresql.org/wiki/Audit_trigger
CREATE SCHEMA IF NOT EXISTS audit;
REVOKE ALL ON SCHEMA audit FROM public;

CREATE TABLE audit.logged_actions (
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

REVOKE ALL ON audit.logged_actions FROM public;

CREATE OR REPLACE FUNCTION audit.if_modified_func() RETURNS trigger AS $body$
DECLARE
    audit_row audit.logged_actions;
    include_values boolean;
    log_queries boolean;
    v_user text;
BEGIN
    IF TG_WHEN <> 'AFTER' THEN
        RAISE EXCEPTION 'audit.if_modified_func() must be fired as an AFTER trigger';
    END IF;

    -- Search for app.current_user in session data
    v_user := current_setting('app.current_user', true);
    IF v_user IS NULL OR v_user = '' THEN
        v_user := session_user;
    END IF;

    audit_row = ROW(
        TG_TABLE_SCHEMA::text,                          -- schema_name
        TG_TABLE_NAME::text,                            -- table_name
        v_user,                                         -- user_name
        clock_timestamp(),                              -- action_tstamp_clk
        txid_current(),                                 -- transaction_id
        substring(TG_OP,1,1),                           -- action (I/U/D/T)
        NULL, NULL,                                     -- original_data, new_data
        current_query()                                 -- query
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
