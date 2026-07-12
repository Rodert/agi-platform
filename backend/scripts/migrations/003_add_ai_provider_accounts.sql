SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `ai_provider_accounts` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(100) NOT NULL,
  `provider` VARCHAR(50) NOT NULL,
  `api_url` VARCHAR(500) NOT NULL,
  `api_key` VARCHAR(500) NOT NULL,
  `extra_config` JSON,
  `is_active` TINYINT(1) NOT NULL DEFAULT 1,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_provider_active` (`provider`, `is_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI供应商账号';

ALTER TABLE `ai_models` ADD COLUMN `provider_account_id` BIGINT NULL AFTER `provider`;
ALTER TABLE `ai_models` ADD KEY `idx_provider_account_id` (`provider_account_id`);

INSERT INTO `ai_provider_accounts` (`name`,`provider`,`api_url`,`api_key`,`extra_config`,`is_active`)
SELECT CONCAT(UPPER(provider),' 默认账号'), provider,
       MAX(COALESCE(JSON_UNQUOTE(JSON_EXTRACT(api_config,'$.api_url')),'')),
       MAX(COALESCE(JSON_UNQUOTE(JSON_EXTRACT(api_config,'$.api_key')),'')),
       JSON_OBJECT(), 1
FROM `ai_models` GROUP BY provider;

UPDATE `ai_models` m JOIN `ai_provider_accounts` a ON a.provider=m.provider
SET m.provider_account_id=a.id WHERE m.provider_account_id IS NULL;
