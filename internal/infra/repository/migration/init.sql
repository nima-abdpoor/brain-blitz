CREATE TABLE user
(
    `id`           bigint primary key AUTO_INCREMENT PRIMARY KEY,
    `username`     varchar(100) NOT NULL,
    `password`     varchar(255) NOT NULL,
    `display_name` varchar(100) NOT NULL,
    `created_at`   TIMESTAMP,
    `updated_at`   TIMESTAMP
);