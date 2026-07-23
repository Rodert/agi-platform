-- Long-lived objects are created only after an administrator approves a work.
INSERT INTO `resource_policies` (`resource_type`,`key_prefix`,`retention_days`,`is_public`,`cache_max_age`,`max_size_mb`) VALUES
('published_image', 'published/images/', 0, 1, 86400, 20),
('published_video', 'published/videos/', 0, 1, 86400, 1024),
('published_thumbnail', 'published/thumbnails/', 0, 1, 86400, 10)
ON DUPLICATE KEY UPDATE `resource_type` = VALUES(`resource_type`);
