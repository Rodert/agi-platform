-- GPT Image 2 supports a richer set of ratio/resolution combinations.
UPDATE `ai_models`
SET `params_config` = JSON_SET(
  `params_config`,
  '$.ratio.options', JSON_ARRAY(
    JSON_OBJECT('value','1:1','label','1:1'), JSON_OBJECT('value','4:3','label','4:3'),
    JSON_OBJECT('value','3:4','label','3:4'), JSON_OBJECT('value','3:2','label','3:2'),
    JSON_OBJECT('value','2:3','label','2:3'), JSON_OBJECT('value','5:4','label','5:4'),
    JSON_OBJECT('value','4:5','label','4:5'), JSON_OBJECT('value','16:9','label','16:9'),
    JSON_OBJECT('value','9:16','label','9:16'), JSON_OBJECT('value','2:1','label','2:1'),
    JSON_OBJECT('value','1:2','label','1:2'), JSON_OBJECT('value','21:9','label','21:9'),
    JSON_OBJECT('value','9:21','label','9:21'), JSON_OBJECT('value','3:1','label','3:1'),
    JSON_OBJECT('value','1:3','label','1:3')
  ),
  '$.ratio.default', '1:1',
  '$.resolution.options', JSON_ARRAY(
    JSON_OBJECT('value','1K','label','1K'), JSON_OBJECT('value','2K','label','2K','extra_cost',1),
    JSON_OBJECT('value','4K','label','4K','extra_cost',2)
  ),
  '$.resolution.default', '1K'
)
WHERE `name` = 'gpt-image-2';
