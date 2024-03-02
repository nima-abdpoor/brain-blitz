CREATE TABLE user
(
    "id"           bigserial NOT NULL AUTO_INCREMENT PRIMARY KEY,
    "username"     varchar   NOT NULL,
    "password"     varchar   NOT NULL,
    "display_name" varchar   NOT NULL,
    "created_at"   TIMESTAMP,
    "updated_at"   TIMESTAMP
);