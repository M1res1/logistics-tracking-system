CREATE TABLE IF NOT EXISTS orders (
    id               SERIAL PRIMARY KEY,
    customer_id      INTEGER NOT NULL,
    restaurant_id    INTEGER NOT NULL,
    driver_id        INTEGER,
    status           VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    subtotal         NUMERIC(10,2) NOT NULL,
    tax              NUMERIC(10,2) NOT NULL,
    delivery_fee     NUMERIC(10,2) NOT NULL,
    total            NUMERIC(10,2) NOT NULL,
    delivery_address TEXT NOT NULL,
    delivery_lat     DOUBLE PRECISION NOT NULL DEFAULT 0,
    delivery_lng     DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orders_customer_id    ON orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_orders_restaurant_id  ON orders(restaurant_id);
CREATE INDEX IF NOT EXISTS idx_orders_status         ON orders(status);
