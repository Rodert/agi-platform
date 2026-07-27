-- Grok Video 1.5 Fast supports only 6- and 10-second videos.
UPDATE `ai_models`
SET `params_config` = JSON_SET(
  `params_config`,
  '$.duration', JSON_OBJECT('label', '视频时长', 'type', 'select', 'default', '6', 'options', JSON_ARRAY(
    JSON_OBJECT('value', '6', 'label', '6 秒'),
    JSON_OBJECT('value', '10', 'label', '10 秒')
  ))
)
WHERE `name` = 'grok-video-1.5fast';
