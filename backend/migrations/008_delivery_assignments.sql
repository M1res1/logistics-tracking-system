CREATE TABLE IF NOT EXISTS delivery_assignments (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    driver_id BIGINT NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING',
    pickup_time TIMESTAMP,
    delivery_time TIMESTAMP,
    distance_km NUMERIC(8,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_da_order_id ON delivery_assignments(order_id);
CREATE INDEX IF NOT EXISTS idx_da_driver_id ON delivery_assignments(driver_id);
CREATE INDEX IF NOT EXISTS idx_da_status ON delivery_assignments(status);

CREATE TABLE IF NOT EXISTS delivery_locations (
    id BIGSERIAL PRIMARY KEY,
    delivery_id BIGINT NOT NULL REFERENCES delivery_assignments(id),
    lat NUMERIC(10,7) NOT NULL,
    lng NUMERIC(10,7) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dl_delivery_id ON delivery_locations(delivery_id);

CREATE TABLE IF NOT EXISTS driver_statuses (
    id BIGSERIAL PRIMARY KEY,
    driver_id BIGINT NOT NULL UNIQUE,
    is_online BOOLEAN NOT NULL DEFAULT false,
    is_available BOOLEAN NOT NULL DEFAULT false,
    last_lat NUMERIC(10,7) NOT NULL DEFAULT 0,
    last_lng NUMERIC(10,7) NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
