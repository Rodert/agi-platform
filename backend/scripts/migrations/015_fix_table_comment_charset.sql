-- Repair table comments that were stored with a mismatched connection charset
-- during earlier initialization. Table names and data are unaffected.
ALTER TABLE `ai_provider_accounts` COMMENT = 'AI供应商账号';
ALTER TABLE `channel_models` COMMENT = '渠道可调用模型绑定';
ALTER TABLE `media_assets` COMMENT = '对象存储资源记录';
ALTER TABLE `prompt_optimization_configs` COMMENT = '提示词优化配置';
ALTER TABLE `prompt_optimization_logs` COMMENT = '提示词优化记录';
ALTER TABLE `resource_policies` COMMENT = '对象存储资源策略';
ALTER TABLE `storage_configs` COMMENT = '存储配置表';
ALTER TABLE `task_attempts` COMMENT = '生成任务执行与重试记录';
ALTER TABLE `task_configs` COMMENT = '任务执行配置';
ALTER TABLE `user_sessions` COMMENT = '用户登录会话';
