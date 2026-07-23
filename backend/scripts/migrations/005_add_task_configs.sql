-- Task execution policy is intentionally isolated from generic system configs.
CREATE TABLE IF NOT EXISTS `task_configs` (
  `id` BIGINT NOT NULL DEFAULT 1,
  `max_active_tasks` INT NOT NULL DEFAULT 50 COMMENT '单用户排队或执行中的任务上限',
  `prompt_max_length` INT NOT NULL DEFAULT 5000 COMMENT '提示词最大字符数',
  `max_retry_attempts` INT NOT NULL DEFAULT 0 COMMENT '任务失败后的最大重试次数',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  CONSTRAINT `chk_task_configs_singleton` CHECK (`id` = 1),
  CONSTRAINT `chk_task_configs_active` CHECK (`max_active_tasks` > 0),
  CONSTRAINT `chk_task_configs_prompt` CHECK (`prompt_max_length` > 0),
  CONSTRAINT `chk_task_configs_retry` CHECK (`max_retry_attempts` >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务执行配置';

INSERT INTO `task_configs` (`id`, `max_active_tasks`, `prompt_max_length`, `max_retry_attempts`)
VALUES (1, 50, 5000, 0)
ON DUPLICATE KEY UPDATE `id` = `id`;
