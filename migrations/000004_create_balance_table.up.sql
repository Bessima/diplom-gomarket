CREATE TABLE IF NOT EXISTS balance
(
    user_id     INT UNIQUE NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    current     BIGINT DEFAULT 0 NOT NULL CHECK (current >= 0),
    withdrawals BIGINT DEFAULT 0 NOT NULL
);