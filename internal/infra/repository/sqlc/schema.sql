CREATE TABLE users
(
    id           BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    username     text   NOT NULL,
    password     text   NOT NULL,
    display_name text   NOT NULL,
    role         text   NOT NULL,
    created_at   TIMESTAMP,
    updated_at   TIMESTAMP
);

CREATE TABLE `permissions`
(
    `id`         INT PRIMARY KEY AUTO_INCREMENT,
    `title`      VARCHAR(255) NOT NULL UNIQUE,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE `access_controls`
(
    `id`            INT PRIMARY KEY AUTO_INCREMENT,
    `actor_id`      INT NOT NULL,
    `actor_type`    ENUM('role', 'user') NOT NULL,
    `permission_id` INT NOT NULL,
    `created_at`    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`permission_id`) REFERENCES `permissions` (`id`)
);