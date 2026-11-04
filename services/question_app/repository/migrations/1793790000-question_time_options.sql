-- +migrate Up
CREATE TABLE question_time_options (
    id VARCHAR(32) PRIMARY KEY,
    label VARCHAR(50) NOT NULL CHECK (length(trim(label)) > 0),
    seconds INTEGER NOT NULL CHECK (seconds BETWEEN 5 AND 3600),
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 0 CHECK (sort_order >= 0)
);
INSERT INTO question_time_options (id, label, seconds, enabled, sort_order)
VALUES ('10-seconds', '10 seconds', 10, TRUE, 10),
       ('1-minute', '1 minute', 60, TRUE, 20)
ON CONFLICT (id) DO NOTHING;

-- +migrate Down
DROP TABLE IF EXISTS question_time_options;
