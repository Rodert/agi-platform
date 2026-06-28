USE `agi_platform`;

UPDATE `video_tasks`
SET
  `images` = COALESCE(`images`, JSON_ARRAY()),
  `videos` = COALESCE(`videos`, JSON_ARRAY()),
  `audios` = COALESCE(`audios`, JSON_ARRAY());
