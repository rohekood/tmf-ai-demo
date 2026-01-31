CREATE TABLE carts (
    id UUID PRIMARY KEY,
    customer_id UUID,
    status VARCHAR(50) NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    total_price_amount DECIMAL(10, 2),
    total_price_currency CHAR(3),
    valid_for_end TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE cart_items (
    id UUID PRIMARY KEY,
    cart_id UUID NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    offering_id UUID NOT NULL,
    quantity INTEGER NOT NULL,
    product_config JSONB,
    unit_amount DECIMAL(10, 2),
    currency CHAR(3),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE outbox_events (
    id UUID PRIMARY KEY,
    topic VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(50) DEFAULT 'PENDING',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_outbox_status ON outbox_events(status);

CREATE TABLE product_prices (
    id UUID PRIMARY KEY, -- Offering ID
    unit_amount DECIMAL(10, 2),
    currency CHAR(3),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
