ALTER TABLE `users`
  ADD COLUMN `is_active` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用' AFTER `level`,
  ADD KEY `idx_user_is_active` (`is_active`);
