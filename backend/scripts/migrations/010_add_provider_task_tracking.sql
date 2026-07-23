-- Preserve the upstream task identifier and its latest normalized state so an
-- asynchronous generation can resume after a Worker restart.
ALTER TABLE `tasks`
  ADD COLUMN `provider_task_id` VARCHAR(255) NOT NULL DEFAULT '' AFTER `channel_id`,
  ADD COLUMN `provider_status` VARCHAR(50) NOT NULL DEFAULT '' AFTER `provider_task_id`,
  ADD COLUMN `provider_response` JSON NULL AFTER `provider_status`,
  ADD COLUMN `last_polled_at` DATETIME NULL AFTER `provider_response`,
  ADD KEY `idx_provider_task_id` (`provider_task_id`),
  ADD KEY `idx_last_polled_at` (`last_polled_at`);
