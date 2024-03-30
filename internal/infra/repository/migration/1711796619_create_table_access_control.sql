-- +migrate Up
CREATE TABLE `access_controls` (
    `id` INT PRIMARY KEY AUTO_INCREMENT,
    `actor_id` INT NOT NULL,
    `actor_type` ENUM('role', 'user') NOT NULL,
    `permission_id` INT NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`permission_id`) REFERENCES `permissions`(`id`)
);

-- +migrate Down
DROP TABLE `access_controls`;