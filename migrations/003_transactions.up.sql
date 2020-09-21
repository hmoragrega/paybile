CREATE TYPE transaction_type AS ENUM ('deposit', 'transfer');

CREATE TABLE IF NOT EXISTS transactions
(
    id               UUID        DEFAULT uuid_generate_v4(),
    wallet_id        UUID,
    amount           NUMERIC(19, 4),
    balance          NUMERIC(19, 4), -- balance after the transaction.
    date             TIMESTAMPTZ DEFAULT NOW(),
    transaction_type transaction_type,
    reference_id     UUID NULL   DEFAULT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (wallet_id) REFERENCES wallets (id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX transactions_date_idx ON transactions USING btree (date, id);
