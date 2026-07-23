-- Storage is a global routing decision: exactly zero or one configuration may
-- be enabled. Preserve the most recently updated legacy configuration first.
UPDATE `storage_configs`
SET `is_enabled` = 0
WHERE `is_enabled` = 1
  AND `id` <> (
    SELECT `selected_id`
    FROM (
      SELECT `id` AS `selected_id`
      FROM `storage_configs`
      WHERE `is_enabled` = 1
      ORDER BY `updated_at` DESC, `id` DESC
      LIMIT 1
    ) AS `selected_storage`
  );

-- MySQL has no partial unique index. NULL values remain repeatable while the
-- generated value 1 can exist only once, enforcing a single active row.
ALTER TABLE `storage_configs`
  ADD COLUMN `active_storage_slot` TINYINT
    GENERATED ALWAYS AS (IF(`is_enabled` = 1, 1, NULL)) STORED,
  ADD UNIQUE KEY `idx_only_one_active_storage` (`active_storage_slot`);
