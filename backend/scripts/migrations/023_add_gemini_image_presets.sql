-- Gemini 3 Pro Image supports all listed aspect ratios and 1K, 2K, 4K output sizes.
UPDATE `ai_models`
SET `params_config` = JSON_OBJECT(
  'ratio', JSON_OBJECT(
    'label', '画面比例', 'type', 'select', 'default', '1:1',
    'options', JSON_ARRAY(
      JSON_OBJECT('value','1:1','label','1:1'), JSON_OBJECT('value','2:3','label','2:3'),
      JSON_OBJECT('value','3:2','label','3:2'), JSON_OBJECT('value','3:4','label','3:4'),
      JSON_OBJECT('value','4:3','label','4:3'), JSON_OBJECT('value','4:5','label','4:5'),
      JSON_OBJECT('value','5:4','label','5:4'), JSON_OBJECT('value','9:16','label','9:16'),
      JSON_OBJECT('value','16:9','label','16:9'), JSON_OBJECT('value','21:9','label','21:9')
    )
  ),
  'resolution', JSON_OBJECT(
    'label', '清晰度', 'type', 'select', 'default', '1K',
    'options', JSON_ARRAY(
      JSON_OBJECT('value','1K','label','1K'), JSON_OBJECT('value','2K','label','2K 高清','extra_cost',1),
      JSON_OBJECT('value','4K','label','4K 超清','extra_cost',2)
    )
  )
)
WHERE `provider` = 'gemini' AND `name` LIKE '%pro-image%';

-- Gemini Flash Image produces 1K images and supports these aspect ratios.
UPDATE `ai_models`
SET `params_config` = JSON_OBJECT(
  'ratio', JSON_OBJECT(
    'label', '画面比例', 'type', 'select', 'default', '1:1',
    'options', JSON_ARRAY(
      JSON_OBJECT('value','1:1','label','1:1'), JSON_OBJECT('value','2:3','label','2:3'),
      JSON_OBJECT('value','3:2','label','3:2'), JSON_OBJECT('value','3:4','label','3:4'),
      JSON_OBJECT('value','4:3','label','4:3'), JSON_OBJECT('value','4:5','label','4:5'),
      JSON_OBJECT('value','5:4','label','5:4'), JSON_OBJECT('value','9:16','label','9:16'),
      JSON_OBJECT('value','16:9','label','16:9')
    )
  )
)
WHERE `provider` = 'gemini' AND `name` LIKE '%flash-image%';
