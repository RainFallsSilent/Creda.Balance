CREATE TABLE balance
(
    id            serial PRIMARY KEY,
    coin_name     text       NOT NULL,
    address       text       NOT NULL,
    balance       text       NOT NULL
);

CREATE TABLE price
(
    id               serial PRIMARY KEY,
    coin_name        text       NOT NULL,
    price            text       NOT NULL,
    timestamp        integer    NOT NULL
);
