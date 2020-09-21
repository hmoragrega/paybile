CREATE TABLE IF NOT EXISTS wallets
(
    id      UUID           DEFAULT uuid_generate_v4(),
    user_id UUID,
    balance NUMERIC(19, 4) DEFAULT 0, -- current balance of the wallet
    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE
);
