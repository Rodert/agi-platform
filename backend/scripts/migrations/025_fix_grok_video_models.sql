-- grok-image-video is an image-to-video model, not an image model.
UPDATE `ai_models`
SET `type` = 'video',
    `params_config` = JSON_OBJECT(
      'ratio', JSON_OBJECT('label', '画面比例', 'type', 'select', 'default', '1:1', 'options', JSON_ARRAY(
        JSON_OBJECT('value','1:1','label','1:1'), JSON_OBJECT('value','16:9','label','16:9'),
        JSON_OBJECT('value','9:16','label','9:16'), JSON_OBJECT('value','4:3','label','4:3'),
        JSON_OBJECT('value','3:4','label','3:4'), JSON_OBJECT('value','3:2','label','3:2'),
        JSON_OBJECT('value','2:3','label','2:3')
      )),
      'resolution', JSON_OBJECT('label', '清晰度', 'type', 'select', 'default', '720p', 'options', JSON_ARRAY(
        JSON_OBJECT('value','720p','label','720p'), JSON_OBJECT('value','480p','label','480p')
      )),
      'duration', JSON_OBJECT('label', '视频时长', 'type', 'select', 'default', '6', 'options', JSON_ARRAY(
        JSON_OBJECT('value','6','label','6 秒'), JSON_OBJECT('value','10','label','10 秒'), JSON_OBJECT('value','15','label','15 秒')
      ))
    )
WHERE `name` = 'grok-image-video';

-- Grok Video 1.5 Fast supports only 6- and 10-second videos.
UPDATE `ai_models`
SET `type` = 'video',
    `params_config` = JSON_OBJECT(
      'ratio', JSON_OBJECT('label', '画面比例', 'type', 'select', 'default', '16:9', 'options', JSON_ARRAY(
        JSON_OBJECT('value','16:9','label','16:9'), JSON_OBJECT('value','9:16','label','9:16')
      )),
      'resolution', JSON_OBJECT('label', '清晰度', 'type', 'select', 'default', '720p', 'options', JSON_ARRAY(
        JSON_OBJECT('value','720p','label','720p'), JSON_OBJECT('value','480p','label','480p')
      )),
      'duration', JSON_OBJECT('label', '视频时长', 'type', 'select', 'default', '6', 'options', JSON_ARRAY(
        JSON_OBJECT('value','6','label','6 秒'), JSON_OBJECT('value','10','label','10 秒')
      ))
    )
WHERE `name` = 'grok-video-1.5fast';
