CREATE TABLE `tokens_` (`id` BIGINT(20) NOT NULL AUTO_INCREMENT, PRIMARY KEY(`id`), `key` CHAR(64) NOT NULL, `resource` CHAR(32) NOT NULL, `data` TEXT, `touched_at` DATETIME NOT NULL, `expire_at` DATETIME NOT NULL, `created_at` DATETIME NOT NULL) DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;
