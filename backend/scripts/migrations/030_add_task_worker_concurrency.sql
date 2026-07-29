ALTER TABLE `task_configs`
  ADD COLUMN `image_concurrency` INT NOT NULL DEFAULT 8 AFTER `max_retry_attempts`,
  ADD COLUMN `video_concurrency` INT NOT NULL DEFAULT 2 AFTER `image_concurrency`;
