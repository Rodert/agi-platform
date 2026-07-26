-- Move the original new-user gift default to user defaults and lower it to 5.
-- The value is only changed when it is still the legacy seed default of 120.
UPDATE `system_configs`
SET `value` = '5', `type` = 'int', `category` = 'user_defaults', `description` = '新用户注册礼包灵感值'
WHERE `key` = 'new_user_gift_amount' AND `value` = '120';

INSERT INTO `system_configs` (`key`, `value`, `type`, `category`, `description`, `updated_at`)
VALUES
  ('default_user_level', 'free', 'string', 'user_defaults', '新用户默认等级', NOW()),
  ('default_user_avatar', '', 'string', 'user_defaults', '新用户默认头像', NOW()),
  ('register_email_verification', 'true', 'bool', 'user_defaults', '注册时是否需要邮箱验证码', NOW())
ON DUPLICATE KEY UPDATE `key` = `key`;
