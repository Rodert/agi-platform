-- Channel accounts are independent from the global model catalog.
-- A model's capability schema is stored once in ai_models; channel_models only controls availability.

ALTER TABLE `ai_provider_accounts`
  ADD COLUMN `priority` INT NOT NULL DEFAULT 100 AFTER `is_active`,
  ADD COLUMN `health_status` VARCHAR(20) NOT NULL DEFAULT 'unknown' AFTER `priority`;

ALTER TABLE `ai_models` ADD UNIQUE KEY `idx_ai_models_name` (`name`);

CREATE TABLE IF NOT EXISTS `channel_models` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `channel_id` BIGINT NOT NULL,
  `model_id` BIGINT NOT NULL,
  `is_active` TINYINT(1) NOT NULL DEFAULT 1,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_channel_model` (`channel_id`, `model_id`),
  KEY `idx_model_active` (`model_id`, `is_active`),
  CONSTRAINT `fk_channel_models_channel` FOREIGN KEY (`channel_id`) REFERENCES `ai_provider_accounts` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_channel_models_model` FOREIGN KEY (`model_id`) REFERENCES `ai_models` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='渠道可调用模型绑定';

-- Preserve existing single-account assignments as channel bindings.
INSERT IGNORE INTO `channel_models` (`channel_id`, `model_id`, `is_active`)
SELECT `provider_account_id`, `id`, `is_active`
FROM `ai_models`
WHERE `provider_account_id` IS NOT NULL;

ALTER TABLE `tasks` ADD COLUMN `channel_id` BIGINT NULL AFTER `model_name`;
ALTER TABLE `tasks` ADD KEY `idx_channel_id` (`channel_id`);
