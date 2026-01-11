CREATE TABLE IF NOT EXISTS catalogs (
    id VARCHAR PRIMARY KEY,
    name VARCHAR NOT NULL,
    description TEXT,
    valid_for_start TIMESTAMP,
    valid_for_end TIMESTAMP,
    last_update TIMESTAMP,
    lifecycle_status VARCHAR
);

CREATE TABLE IF NOT EXISTS categories (
    id VARCHAR PRIMARY KEY,
    name VARCHAR NOT NULL,
    description TEXT,
    parent_id VARCHAR,
    is_root BOOLEAN,
    catalog_id VARCHAR,
    valid_for_start TIMESTAMP,
    valid_for_end TIMESTAMP,
    last_update TIMESTAMP,
    lifecycle_status VARCHAR,
    CONSTRAINT fk_catalog FOREIGN KEY (catalog_id) REFERENCES catalogs(id)
);

CREATE TABLE IF NOT EXISTS product_specifications (
    id VARCHAR PRIMARY KEY,
    name VARCHAR NOT NULL,
    description TEXT,
    product_number VARCHAR,
    lifecycle_status VARCHAR,
    valid_for_start TIMESTAMP,
    valid_for_end TIMESTAMP,
    last_update TIMESTAMP,
    characteristics JSONB
);

CREATE TABLE IF NOT EXISTS product_offerings (
    id VARCHAR PRIMARY KEY,
    name VARCHAR NOT NULL,
    description TEXT,
    lifecycle_status VARCHAR,
    valid_for_start TIMESTAMP,
    valid_for_end TIMESTAMP,
    last_update TIMESTAMP,
    is_bundle BOOLEAN,
    is_sellable BOOLEAN,
    product_specification_id VARCHAR,
    product_offering_price JSONB,
    category_ids JSONB,
    attachments JSONB,
    CONSTRAINT fk_specification FOREIGN KEY (product_specification_id) REFERENCES product_specifications(id)
);

CREATE TABLE IF NOT EXISTS outbox_events (
    id UUID PRIMARY KEY,
    routing_key VARCHAR NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_outbox_events_status ON outbox_events(status);
