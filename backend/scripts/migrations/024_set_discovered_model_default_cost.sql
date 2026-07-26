-- Models created by channel discovery default to 100 inspiration credits.
-- Preserve any non-zero amount already configured by an administrator.
UPDATE `ai_models`
SET `cost` = 100
WHERE `cost` = 0;
