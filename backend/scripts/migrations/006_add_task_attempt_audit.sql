ALTER TABLE `tasks`
  ADD COLUMN `attempt_count` INT NOT NULL DEFAULT 0 AFTER `cost`,
  ADD COLUMN `max_retry_attempts` INT NOT NULL DEFAULT 0 AFTER `attempt_count`,
  ADD COLUMN `last_retry_at` DATETIME NULL AFTER `max_retry_attempts`,
  ADD KEY `idx_last_retry_at` (`last_retry_at`);

CREATE TABLE IF NOT EXISTS `task_attempts` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `task_id` BIGINT NOT NULL,
  `attempt` INT NOT NULL,
  `status` VARCHAR(20) NOT NULL COMMENT 'processing/success/failed',
  `error_msg` VARCHAR(1000) NOT NULL DEFAULT '',
  `started_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `completed_at` DATETIME NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_task_attempt` (`task_id`, `attempt`),
  KEY `idx_task_attempt_status` (`task_id`, `status`),
  CONSTRAINT `fk_task_attempts_task` FOREIGN KEY (`task_id`) REFERENCES `tasks` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='生成任务执行与重试记录';

-- Existing completed or in-progress tasks predate attempt auditing. Preserve
-- their known single execution without inventing retry history.
UPDATE `tasks`
SET `attempt_count` = 1
WHERE `attempt_count` = 0 AND `status` IN ('processing', 'success', 'failed');

INSERT IGNORE INTO `task_attempts` (`task_id`, `attempt`, `status`, `error_msg`, `started_at`, `completed_at`)
SELECT `id`, 1, `status`, COALESCE(`error_msg`, ''), `created_at`, `completed_at`
FROM `tasks`
WHERE `attempt_count` = 1;
