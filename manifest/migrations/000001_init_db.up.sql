CREATE TABLE balance
(
    id            serial PRIMARY KEY,
    timestamp     integer    NOT NULL,
    address       text       NOT NULL,
    -- coin_id       text       NOT NULL,
    -- balance       text       NOT NULL,
    total_usd     text       NOT NULL
);


CREATE TABLE price
(
    id               serial PRIMARY KEY,
    coin_id          text       NOT NULL,
    price            text       NOT NULL,
    timestamp        integer    NOT NULL
);
