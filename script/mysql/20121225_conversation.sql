CREATE TABLE `posts` (`id` BIGINT(20) NOT NULL AUTO_INCREMENT, `by_id` BIGINT(20) NOT NULL, `created_at` DATETIME NOT NULL, `relationship` TEXT, `content` TEXT, `via` VARCHAR(255), `exfee_id` BIGINT(20), `ref_uri` VARCHAR(255)) DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;