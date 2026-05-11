CREATE TABLE IF NOT EXISTS order_items (
    id                   SERIAL PRIMARY KEY,
    order_id             INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    menu_item_id         INTEGER NOT NULL,
    quantity             INTEGER NOT NULL CHECK (quantity > 0),
    unit_price           NUMERIC(10,2) NOT NULL,
    special_instructions TEXT
);

CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
