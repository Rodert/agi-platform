-- Generated images, videos and their thumbnails are temporary user assets.
-- Published copies keep their separate published_* policies and remain unaffected.
UPDATE `resource_policies`
SET `retention_days` = 7
WHERE `resource_type` IN ('image', 'video', 'thumbnail');

-- Existing temporary assets created before this migration did not always have
-- an expiry. Bring them under the same seven-day policy without touching
-- independent published_* assets.
UPDATE `media_assets`
SET `expires_at` = DATE_ADD(`created_at`, INTERVAL 7 DAY)
WHERE `resource_type` IN ('image', 'video', 'thumbnail')
  AND `expires_at` IS NULL;
