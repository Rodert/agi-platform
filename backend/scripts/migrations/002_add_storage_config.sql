-- 添加存储配置表
CREATE TABLE IF NOT EXISTS `storage_configs` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL COMMENT '配置名称',
  `type` varchar(50) NOT NULL COMMENT '存储类型: local, tencent_cos, aliyun_oss, cloudflare',
  `local_path` varchar(255) DEFAULT NULL COMMENT '本地存储路径',
  `endpoint` varchar(255) DEFAULT NULL COMMENT '端点',
  `access_key` varchar(255) DEFAULT NULL COMMENT 'AccessKey',
  `secret_key` varchar(255) DEFAULT NULL COMMENT 'SecretKey',
  `bucket` varchar(100) DEFAULT NULL COMMENT '桶名称',
  `region` varchar(50) DEFAULT NULL COMMENT '区域',
  `domain` varchar(255) DEFAULT NULL COMMENT 'CDN域名',
  `is_enabled` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否启用',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_type` (`type`),
  KEY `idx_is_enabled` (`is_enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='存储配置表';

-- 插入默认的本地存储配置
INSERT INTO `storage_configs` (`name`, `type`, `local_path`, `domain`, `is_enabled`)
VALUES ('本地存储', 'local', './uploads', 'http://localhost:8080/uploads', 1);
