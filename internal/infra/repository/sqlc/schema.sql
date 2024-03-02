CREATE TABLE user
(
    id           BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    username     text   NOT NULL,
    password     text   NOT NULL,
    display_name text   NOT NULL,
    created_at   TIMESTAMP,
    updated_at   TIMESTAMP
);