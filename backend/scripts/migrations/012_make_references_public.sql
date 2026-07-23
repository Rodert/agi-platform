UPDATE `resource_policies`
SET `is_public` = 1, `cache_max_age` = CASE WHEN `cache_max_age` = 0 THEN 86400 ELSE `cache_max_age` END
WHERE `resource_type` = 'reference';
