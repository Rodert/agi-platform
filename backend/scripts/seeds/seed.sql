-- AGI Platform 种子数据
-- 初始化配置数据

SET NAMES utf8mb4;

-- ----------------------------
-- 1. 系统配置
-- ----------------------------
INSERT INTO `system_configs` (`key`, `value`, `type`, `category`, `description`) VALUES
-- 积分配置
('new_user_gift_amount', '5', 'int', 'user_defaults', '新用户注册礼包灵感值'),
('default_user_level', 'free', 'string', 'user_defaults', '新用户默认等级'),
('default_user_avatar', '', 'string', 'user_defaults', '新用户默认头像'),
('register_email_verification', 'true', 'bool', 'user_defaults', '注册时是否需要邮箱验证码'),
('checkin_base_reward', '1', 'int', 'checkin', '签到基础奖励'),
('checkin_7day_reward', '3', 'int', 'checkin', '连续7天签到奖励'),

-- 邀请配置
('invite_register_reward_inviter', '50', 'int', 'invitation', '邀请人注册奖励'),
('invite_register_reward_invitee', '20', 'int', 'invitation', '被邀请人注册奖励'),
('invite_recharge_reward_inviter', '100', 'int', 'invitation', '邀请人首充奖励'),
('invite_recharge_enabled', 'true', 'bool', 'invitation', '是否启用首充奖励'),

-- 审核配置
('audit_keywords', '["暴力", "色情", "政治"]', 'json', 'audit', '审核敏感词')
ON DUPLICATE KEY UPDATE `key` = `key`;

-- ----------------------------
-- 1.1 任务配置
-- ----------------------------
INSERT INTO `task_configs` (`id`, `max_active_tasks`, `prompt_max_length`, `max_retry_attempts`) VALUES
(1, 50, 5000, 0)
ON DUPLICATE KEY UPDATE `id` = `id`;

-- ----------------------------
-- 2. 充值套餐
-- ----------------------------
INSERT INTO `credit_packages` (`name`, `price`, `points`, `note`, `is_hot`, `sort_order`, `is_active`) VALUES
('入门版', 12.90, 100, '约生成 25 张图片', 0, 1, 1),
('畅享版', 55.90, 500, '约生成 125 张图片', 1, 2, 1),
('专业版', 99.90, 1000, '重度创作者首选', 0, 3, 1);

-- ----------------------------
-- 3. AI 模型配置
-- ----------------------------
INSERT INTO `ai_models` (`name`, `display_name`, `type`, `provider`, `description`, `logo_url`, `tag`, `cost`, `api_config`, `params_config`, `is_active`, `sort_order`) VALUES
-- 图片模型
('gpt-image-2', 'GPT Image 2', 'image', 'openai', '新一代图像模型，中文支持优秀', '/logos/gpt.png', '推荐', 4,
 '{"api_url":"https://api.openai.com/v1/images/generations","model":"dall-e-3","timeout":120}',
 '{"ratio":{"label":"画面比例","type":"select","options":[{"value":"1:1","label":"1:1 方形"},{"value":"16:9","label":"16:9 横版"},{"value":"9:16","label":"9:16 竖版"}],"default":"1:1"},"resolution":{"label":"分辨率","type":"select","options":[{"value":"1K","label":"1K"},{"value":"2K","label":"2K","extra_cost":2}],"default":"1K"}}',
 1, 1),

('seedream-4.0', 'Seedream 4.0', 'image', 'jimeng', '中文友好的图像生成模型', '/logos/jm.png', '中文友好', 4,
 '{"api_url":"https://api.jimeng.ai/v1/generate","timeout":60}',
 '{"ratio":{"label":"画面比例","type":"select","options":[{"value":"1:1","label":"1:1"},{"value":"16:9","label":"16:9"}],"default":"1:1"}}',
 1, 2),

-- 视频模型
('wave-1.5', 'Wave 1.5', 'video', 'wave', '最新一代视频生成模型', NULL, '新品上线', 18,
 '{"api_url":"https://api.wave.ai/v1/video/generate","timeout":300}',
 '{"ratio":{"label":"画面比例","type":"select","options":[{"value":"16:9","label":"16:9"},{"value":"9:16","label":"9:16"}],"default":"16:9"},"resolution":{"label":"清晰度","type":"select","options":[{"value":"720P","label":"720P"},{"value":"1080P","label":"1080P","extra_cost":6}],"default":"720P"},"duration":{"label":"时长","type":"select","options":[{"value":"5s","label":"5秒"},{"value":"10s","label":"10秒","extra_cost":12}],"default":"5s"},"sound":{"label":"生成声音","type":"switch","default":false}}',
 1, 1);

-- ----------------------------
-- 4. 作品分类
-- ----------------------------
INSERT INTO `categories` (`name`, `type`, `sort_order`, `is_active`) VALUES
-- 作品分类
('全部', 'work_category', 0, 1),
('灵感', 'work_category', 1, 1),
('人像', 'work_category', 2, 1),
('创意', 'work_category', 3, 1),
('写真', 'work_category', 4, 1),
('电影', 'work_category', 5, 1),
('视觉', 'work_category', 6, 1),
('时尚', 'work_category', 7, 1),
('3D', 'work_category', 8, 1),
('产品', 'work_category', 9, 1),

-- 画面比例
('1:1', 'aspect_ratio', 1, 1),
('16:9', 'aspect_ratio', 2, 1),
('9:16', 'aspect_ratio', 3, 1),
('4:5', 'aspect_ratio', 4, 1),
('2:3', 'aspect_ratio', 5, 1);

-- ----------------------------
-- 5. 支付渠道（示例）
-- ----------------------------
INSERT INTO `payment_channels` (`name`, `channel_type`, `merchant_id`, `config`, `is_active`, `sort_order`) VALUES
('易支付-主账号', 'epay', 'merchant_001',
 '{"api_url":"https://pay.example.com","partner_id":"10001","partner_key":"your-key","sign_type":"MD5"}',
 0, 1),

('演示支付', 'demo', 'demo',
 '{"description":"测试环境模拟支付，自动完成"}',
 1, 99);

-- ----------------------------
-- 6. 兑换码（示例）
-- ----------------------------
INSERT INTO `redeem_codes` (`code`, `amount`, `batch_id`, `batch_name`, `expires_at`) VALUES
('TIDE2026', 50, 'BATCH001', '测试兑换码', DATE_ADD(NOW(), INTERVAL 30 DAY)),
('WELCOME100', 100, 'BATCH002', '新用户欢迎码', DATE_ADD(NOW(), INTERVAL 30 DAY));

-- ----------------------------
-- 7. 邮箱配置（示例，需要修改）
-- ----------------------------
INSERT INTO `email_config` (`id`, `smtp_host`, `smtp_port`, `smtp_user`, `smtp_password`, `smtp_ssl`, `from_name`, `from_email`, `is_active`) VALUES
(1, 'smtp.gmail.com', 587, 'your-email@gmail.com', 'your-password', 0, '潮汐AI', 'noreply@tide.ai', 0);

-- 提示：请修改以上配置为实际值
