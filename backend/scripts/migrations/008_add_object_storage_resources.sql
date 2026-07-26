CREATE TABLE IF NOT EXISTS `resource_policies` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `resource_type` VARCHAR(30) NOT NULL,
  `key_prefix` VARCHAR(100) NOT NULL,
  `retention_days` INT NOT NULL DEFAULT 0 COMMENT '0 means retained until manually deleted',
  `is_public` TINYINT(1) NOT NULL DEFAULT 1,
  `cache_max_age` INT NOT NULL DEFAULT 86400,
  `max_size_mb` INT NOT NULL,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_resource_type` (`resource_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='对象存储资源策略';

INSERT INTO `resource_policies` (`resource_type`,`key_prefix`,`retention_days`,`is_public`,`cache_max_age`,`max_size_mb`) VALUES
('image', 'images/', 7, 1, 86400, 20),
('video', 'videos/', 7, 1, 86400, 1024),
('thumbnail', 'thumbnails/', 7, 1, 86400, 10),
('reference', 'references/', 30, 1, 86400, 5)
ON DUPLICATE KEY UPDATE `resource_type` = VALUES(`resource_type`);

CREATE TABLE IF NOT EXISTS `media_assets` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `task_id` BIGINT NULL,
  `user_id` BIGINT NOT NULL,
  `storage_config_id` BIGINT NOT NULL,
  `resource_type` VARCHAR(30) NOT NULL,
  `object_key` VARCHAR(500) NOT NULL,
  `public_url` VARCHAR(1000) NULL,
  `content_type` VARCHAR(100) NOT NULL,
  `size_bytes` BIGINT NOT NULL,
  `expires_at` DATETIME NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_media_object_key` (`object_key`),
  KEY `idx_media_task` (`task_id`),
  KEY `idx_media_user` (`user_id`),
  KEY `idx_media_type` (`resource_type`),
  KEY `idx_media_expires` (`expires_at`),
  CONSTRAINT `fk_media_assets_task` FOREIGN KEY (`task_id`) REFERENCES `tasks` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_media_assets_storage` FOREIGN KEY (`storage_config_id`) REFERENCES `storage_configs` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='对象存储资源记录';
