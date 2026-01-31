DROP TRIGGER IF EXISTS audit_trigger_prices ON product_prices;
DROP TRIGGER IF EXISTS audit_trigger_cart_items ON cart_items;
DROP TRIGGER IF EXISTS audit_trigger_carts ON carts;
DROP FUNCTION IF EXISTS audit.if_modified_func;
DROP TABLE IF EXISTS audit.logged_actions;
DROP SCHEMA IF EXISTS audit;
