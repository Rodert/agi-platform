-- Persist the narrow interval after an upstream request is sent but before its
-- task ID is returned. On restart these tasks are failed and refunded rather
-- than being submitted again to non-idempotent upstreams.
ALTER TABLE `tasks`
  ADD COLUMN `submission_state` VARCHAR(20) NOT NULL DEFAULT '' AFTER `provider_status`,
  ADD INDEX `idx_tasks_submission_state` (`submission_state`);

-- Tasks already processing during this rollout have no persisted submission
-- boundary. Treat them conservatively on the next Worker start instead of
-- risking a duplicate upstream request.
UPDATE `tasks`
  SET `submission_state` = 'submitting'
  WHERE `status` = 'processing' AND `provider_task_id` = '';
