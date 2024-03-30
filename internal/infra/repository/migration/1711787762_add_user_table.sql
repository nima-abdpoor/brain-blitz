-- +migrate Up
-- please read this article to understand why we use VARCHAR(191)
-- https://www.grouparoo.com/blog/varchar-191#why-varchar-and-not-text
CREATE TABLE `users`
(
    `id`           bigint primary key AUTO_INCREMENT PRIMARY KEY,
    `username`     varchar(100) NOT NULL UNIQUE,
    `password`     varchar(255) NOT NULL,
    `display_name` varchar(100) NOT NULL,
    `created_at`   TIMESTAMP,
    `updated_at`   TIMESTAMP
);


-- +migrate Down
DROP TABLE `users`;