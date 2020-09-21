CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users
(
    id              UUID DEFAULT uuid_generate_v4(),
    login           VARCHAR(256),
    hashed_password VARCHAR(64),
    PRIMARY KEY (id)
);
