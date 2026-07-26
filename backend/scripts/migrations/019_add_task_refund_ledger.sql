-- Link task charges to their task and make failed-task refunds auditable.
-- Existing task expenses were created without a task reference and are not
-- eligible for automatic refunds; newly created tasks use these identifiers.
ALTER TABLE `credit_ledgers`
  ADD KEY `idx_task_expense` (`source_type`, `source_id`);
