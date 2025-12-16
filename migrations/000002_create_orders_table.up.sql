CREATE TYPE order_status AS ENUM ('NEW', 'INVALID', 'PROCESSING', 'PROCESSED');

CREATE TABLE IF NOT EXISTS orders
(
    id          BIGINT PRIMARY KEY,
    user_id     INT                      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    accrual     BIGINT,
    status      order_status             NOT NULL DEFAULT 'NEW',
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);




