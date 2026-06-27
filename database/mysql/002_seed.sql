USE `agi_platform`;

SET NAMES utf8mb4;

INSERT INTO `admin_users` (`id`, `username`, `password_hash`, `nickname`, `role`, `status`)
VALUES (1, 'admin', '$2a$10$No72On/WKoV/8ZRXpFRM5uzIRP5YoXe544GX79XE3bWTFOmglEPRu', 'Administrator', 'super_admin', 'active')
ON DUPLICATE KEY UPDATE `nickname` = VALUES(`nickname`), `role` = VALUES(`role`);

INSERT INTO `users` (`id`, `email`, `password_hash`, `nickname`, `credits`, `status`)
VALUES (1, 'demo@example.com', 'not-set', 'Demo User', 1000, 'active')
ON DUPLICATE KEY UPDATE `nickname` = VALUES(`nickname`);

INSERT INTO `providers` (`id`, `code`, `name`, `type`, `base_url`, `enabled`, `timeout_seconds`, `retry_count`, `priority`, `remark`)
VALUES (1, 'mock', 'Mock Provider', 'mock', '', 1, 60, 1, 100, 'Local development provider')
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`), `type` = VALUES(`type`);

INSERT INTO `image_models` (
  `id`,
  `code`,
  `display_name`,
  `description`,
  `price_credits`,
  `supported_sizes`,
  `support_text_to_image`,
  `max_images_per_request`,
  `auto_refund_on_failure`,
  `enabled`,
  `recommended`,
  `sort_order`
)
VALUES (
  1,
  'general-high-quality',
  '通用高质量',
  '适合通用图片、商品图和海报封面',
  8,
  JSON_ARRAY('1024x1024', '1024x1536', '1536x1024'),
  1,
  4,
  1,
  1,
  1,
  10
)
ON DUPLICATE KEY UPDATE `display_name` = VALUES(`display_name`);

INSERT INTO `image_model_routes` (`id`, `model_id`, `provider_id`, `provider_model_name`, `enabled`, `priority`, `weight`)
VALUES (1, 1, 1, 'mock-image', 1, 100, 100)
ON DUPLICATE KEY UPDATE `provider_model_name` = VALUES(`provider_model_name`);
