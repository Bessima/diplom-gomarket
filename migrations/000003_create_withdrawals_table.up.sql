CREATE TABLE IF NOT EXISTS withdrawals
(
    order_id     BIGINT PRIMARY KEY,
    user_id      INT                      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    sum          BIGINT,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);