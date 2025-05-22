-- +migrate Up
CREATE TABLE users
(
    id       UUID PRIMARY KEY,
    username TEXT UNIQUE NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS users;
