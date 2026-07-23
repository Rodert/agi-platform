CREATE TABLE IF NOT EXISTS `prompt_optimization_configs` (
  `id` BIGINT NOT NULL DEFAULT 1,
  `is_active` TINYINT(1) NOT NULL DEFAULT 0,
  `model_name` VARCHAR(100) NOT NULL DEFAULT '',
  `system_prompt` TEXT NOT NULL,
  `max_input_length` INT NOT NULL DEFAULT 5000,
  `credit_cost` INT NOT NULL DEFAULT 0,
  `rate_limit_per_minute` INT NOT NULL DEFAULT 5,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  CONSTRAINT `chk_prompt_optimization_config_singleton` CHECK (`id` = 1)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='提示词优化配置';

INSERT INTO `prompt_optimization_configs` (`id`, `is_active`, `model_name`, `system_prompt`, `max_input_length`, `credit_cost`, `rate_limit_per_minute`)
VALUES (1, 0, '', 'You improve prompts for AI image and video generation. Preserve the user intent, named entities, language, and explicit constraints. Return only the improved prompt, without commentary, labels, quotation marks, or Markdown. For images, make composition, visual details, lighting, and style concrete. For videos, make subjects, action, camera movement, scene progression, and timing concrete. Do not invent brands, people, or requirements the user did not request.', 5000, 0, 5)
ON DUPLICATE KEY UPDATE `id` = `id`;

CREATE TABLE IF NOT EXISTS `prompt_optimization_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT NOT NULL,
  `model_name` VARCHAR(100) NOT NULL,
  `channel_id` BIGINT NOT NULL,
  `target_type` VARCHAR(20) NOT NULL,
  `target_model_name` VARCHAR(100) NULL,
  `params` JSON NULL,
  `original_prompt` TEXT NOT NULL,
  `optimized_prompt` TEXT NULL,
  `credit_cost` INT NOT NULL DEFAULT 0,
  `status` VARCHAR(20) NOT NULL,
  `error_msg` VARCHAR(1000) NULL,
  `latency_ms` INT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_prompt_optimization_user_created` (`user_id`, `created_at`),
  KEY `idx_prompt_optimization_channel` (`channel_id`),
  KEY `idx_prompt_optimization_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='提示词优化记录';
