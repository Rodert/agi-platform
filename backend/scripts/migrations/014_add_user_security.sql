-- User phone binding and server-side session revocation.
ALTER TABLE `users` ADD COLUMN `phone` VARCHAR(20) NULL AFTER `invited_by`;
ALTER TABLE `users` ADD UNIQUE KEY `idx_phone` (`phone`);

CREATE TABLE `user_sessions` (
  `id` VARCHAR(64) NOT NULL,
  `user_id` BIGINT NOT NULL,
  `device` VARCHAR(255) NOT NULL,
  `ip` VARCHAR(64) NULL,
  `expires_at` DATETIME NOT NULL,
  `revoked_at` DATETIME NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_sessions_user_id` (`user_id`),
  KEY `idx_user_sessions_expires_at` (`expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户登录会话';
