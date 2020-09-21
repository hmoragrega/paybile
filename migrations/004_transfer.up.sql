CREATE TABLE IF NOT EXISTS transfers
(
    id                    UUID        DEFAULT uuid_generate_v4(),
    issuer_id             UUID,
    origin_wallet_id      UUID,
    destination_wallet_id UUID,
    amount                DECIMAL(19, 4),
    message               TEXT,
    date                  TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (id),
    FOREIGN KEY (issuer_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (origin_wallet_id) REFERENCES wallets (id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (destination_wallet_id) REFERENCES wallets (id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX transfers_date_idx ON transfers USING btree (date, id);
