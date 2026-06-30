-- UNUSED: This table is never read or written by any application code.
-- It was created as a mirror of user_app's users table (different DB, different PK type: UUID vs SERIAL).
-- Retained as schema reference; no service depends on it. See issue #12 in docs/context/07-known-issues.md.

-- +migrate Up
CREATE TABLE users
(
    id       UUID PRIMARY KEY,
    username TEXT UNIQUE NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS users;
