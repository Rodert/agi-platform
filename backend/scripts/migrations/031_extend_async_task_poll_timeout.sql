-- Preserve administrator customizations while extending the former default
-- from 15 minutes to one hour for expensive asynchronous video jobs.
UPDATE ai_provider_accounts
SET extra_config = JSON_SET(COALESCE(extra_config, JSON_OBJECT()), '$.poll_timeout_seconds', 3600)
WHERE provider IN ('grok', 'jimeng_international')
  AND CAST(JSON_UNQUOTE(JSON_EXTRACT(extra_config, '$.poll_timeout_seconds')) AS UNSIGNED) = 900;
