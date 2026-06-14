CREATE TABLE `friends` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
    `user_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
    `friend_uid` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
    `remark` varchar(255) DEFAULT NULL,
    `add_source` tinyint COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `created_at` timestamp NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `friend_requests` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
    `user_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
    `req_uid` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL,
    `req_msg` varchar(255) DEFAULT NULL,
    `req_time` timestamp NOT NULL,
    `handle_result` tinyint COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    `handle_msg` varchar(255) DEFAULT NULL,
    `handled_at` timestamp NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;