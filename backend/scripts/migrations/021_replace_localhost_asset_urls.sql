-- Local files are served through the application's reverse proxy. Keep stored
-- URLs relative so they work on every deployment domain.
UPDATE `media_assets`
SET `public_url` = REPLACE(`public_url`, 'http://localhost:8080/uploads/', '/uploads/')
WHERE `public_url` LIKE 'http://localhost:8080/uploads/%';

UPDATE `tasks`
SET `result_url` = REPLACE(`result_url`, 'http://localhost:8080/uploads/', '/uploads/'),
    `thumbnail_url` = REPLACE(`thumbnail_url`, 'http://localhost:8080/uploads/', '/uploads/')
WHERE `result_url` LIKE 'http://localhost:8080/uploads/%'
   OR `thumbnail_url` LIKE 'http://localhost:8080/uploads/%';

UPDATE `works`
SET `image_url` = REPLACE(`image_url`, 'http://localhost:8080/uploads/', '/uploads/'),
    `video_url` = REPLACE(`video_url`, 'http://localhost:8080/uploads/', '/uploads/')
WHERE `image_url` LIKE 'http://localhost:8080/uploads/%'
   OR `video_url` LIKE 'http://localhost:8080/uploads/%';
