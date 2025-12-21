CREATE TABLE IF NOT EXISTS parties (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    href VARCHAR(255),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS individuals (
    id VARCHAR(255) PRIMARY KEY REFERENCES parties(id),
    given_name VARCHAR(255),
    family_name VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS organizations (
    id VARCHAR(255) PRIMARY KEY REFERENCES parties(id),
    trading_name VARCHAR(255),
    is_legal_entity BOOLEAN
);
