-- +migrate Up
CREATE TABLE match_questions
(
    match_id    UUID,
    question_id UUID,
    PRIMARY KEY (match_id, question_id)
);

-- +migrate Down
DROP TABLE IF EXISTS match_questions;
