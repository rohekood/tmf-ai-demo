-- Rollback audit schema
DROP FUNCTION IF EXISTS audit.if_modified_func();
DROP TABLE IF EXISTS audit.logged_actions;
DROP SCHEMA IF EXISTS audit;
