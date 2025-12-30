CREATE TABLE IF NOT EXISTS applied_billing_rates (
    id VARCHAR(255) PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    product_ref VARCHAR(100) NOT NULL,
    rate_type VARCHAR(50) NOT NULL,
    value DECIMAL(10, 2) NOT NULL,
    valid_for_start TIMESTAMP WITH TIME ZONE NOT NULL,
    valid_for_end TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_applied_billing_rates_customer_id ON applied_billing_rates(customer_id);
