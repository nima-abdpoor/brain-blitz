-- UNUSED AT RUNTIME: The table schema is referenced by GetProperQuestions (question deduplication logic),
-- but GetProperQuestions is never called — GetRandomQuestions is always used instead.
-- Nothing writes to this table, so it remains empty. Planned for Phase 3 (R-09).
-- See issue #11 in docs/context/07-known-issues.md.

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
