CREATE TABLE saga_instances (
    id UUID PRIMARY KEY,
    cart_id UUID NOT NULL UNIQUE,
    customer_id UUID,
    current_step VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    payload JSONB,
    history JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE outbox_events (
    id UUID PRIMARY KEY,
    topic VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    headers JSONB,
    status VARCHAR(50) DEFAULT 'PENDING',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP
);

CREATE INDEX idx_outbox_status ON outbox_events(status);
