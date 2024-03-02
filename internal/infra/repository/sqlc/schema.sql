CREATE TABLE user
(
    id           BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    email        text   NOT NULL,
    password     text,
    display_name text,
    created_at   TIMESTAMP,
    updated_at   TIMESTAMP
);