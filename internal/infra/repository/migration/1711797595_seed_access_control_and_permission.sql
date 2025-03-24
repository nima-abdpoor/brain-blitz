-- +migrate Up
INSERT INTO `permissions` (`id`, `title`) VALUES (1, 'USER_LIST');
INSERT INTO `permissions` (`id`, `title`) VALUES (2, 'USER_DELETE');

INSERT INTO `access_controls` (`actor_type`, `actor_id`, `permission_id`) VALUES ('role', 2, 1);
INSERT INTO `access_controls` (`actor_type`, `actor_id`, `permission_id`) VALUES ('role', 2, 2);

-- +migrate Down
DELETE FROM `permissions` WHERE id in (1, 2);