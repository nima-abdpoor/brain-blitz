-- +migrate Up
CREATE TABLE user_question_history
(
    user_id     bigInt,
    question_id UUID,
    seen_at     TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, question_id)
);

-- +migrate Down
DROP TABLE IF EXISTS user_question_history;
