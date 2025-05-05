-- +migrate Up
CREATE TABLE questions
(
    id             UUID PRIMARY KEY,
    content        TEXT NOT NULL,
    correct_answer TEXT NOT NULL,
    choices        TEXT[],
    category       TEXT,
    difficulty     TEXT,
    created_at     TIMESTAMP DEFAULT NOW()
);

-- +migrate Down
DROP TABLE IF EXISTS questions;
