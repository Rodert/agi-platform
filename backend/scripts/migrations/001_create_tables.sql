-- AGI Platform Database Migration
-- 创建数据库表结构

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- 1. 用户模块
-- ----------------------------

-- 用户表
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `email` VARCHAR(100) NOT NULL,
  `password_hash` VARCHAR(255),
  `name` VARCHAR(50) NOT NULL,
  `avatar` VARCHAR(255),
  `bio` VARCHAR(500),
  `level` VARCHAR(20) NOT NULL DEFAULT 'free' COMMENT '会员等级: free/member/pro',
  `invite_code` VARCHAR(10) NOT NULL COMMENT '邀请码',
  `invited_by` BIGINT COMMENT '邀请人ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_email` (`email`),
  UNIQUE KEY `idx_invite_code` (`invite_code`),
  KEY `idx_invited_by` (`invited_by`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- 验证码表
DROP TABLE IF EXISTS `verification_codes`;
CREATE TABLE `verification_codes` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `email` VARCHAR(100) NOT NULL,
  `code` VARCHAR(10) NOT NULL,
  `type` VARCHAR(20) NOT NULL COMMENT 'register/login/reset',
  `expires_at` DATETIME NOT NULL,
  `used_at` DATETIME,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_email_type` (`email`, `type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='验证码表';

-- ----------------------------
-- 2. 创作模块
-- ----------------------------

-- 生成请求表
DROP TABLE IF EXISTS `generation_requests`;
CREATE TABLE `generation_requests` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT NOT NULL,
  `prompt` TEXT NOT NULL,
  `model_name` VARCHAR(50) NOT NULL,
  `type` VARCHAR(20) NOT NULL COMMENT 'image/video/product',
  `params` JSON,
  `cost` INT NOT NULL,
  `task_id` BIGINT,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_task_id` (`task_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='生成请求表';

-- 任务表
DROP TABLE IF EXISTS `tasks`;
CREATE TABLE `tasks` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT NOT NULL,
  `request_id` BIGINT NOT NULL,
  `title` VARCHAR(255) NOT NULL,
  `type` VARCHAR(20) NOT NULL COMMENT 'image/video/product',
  `status` VARCHAR(20) NOT NULL DEFAULT 'queued' COMMENT 'queued/processing/success/failed',
  `progress` INT NOT NULL DEFAULT 0,
  `prompt` TEXT NOT NULL,
  `model_name` VARCHAR(50) NOT NULL,
  `result_url` VARCHAR(500),
  `thumbnail_url` VARCHAR(500),
  `error_msg` VARCHAR(500),
  `cost` INT NOT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `completed_at` DATETIME,
  PRIMARY KEY (`id`),
  KEY `idx_user_status` (`user_id`, `status`),
  KEY `idx_request_id` (`request_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_completed_at` (`completed_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务表';

-- ----------------------------
-- 3. 作品模块
-- ----------------------------

-- 作品表
DROP TABLE IF EXISTS `works`;
CREATE TABLE `works` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT NOT NULL,
  `task_id` BIGINT,
  `title` VARCHAR(255) NOT NULL,
  `prompt` TEXT NOT NULL,
  `category` VARCHAR(50),
  `type` VARCHAR(20) NOT NULL COMMENT 'image/video',
  `ratio` VARCHAR(10),
  `image_url` VARCHAR(500),
  `video_url` VARCHAR(500),
  `audit_status` VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT 'pending/approved/rejected',
  `audit_reason` VARCHAR(500),
  `audit_admin_id` BIGINT,
  `audited_at` DATETIME,
  `likes_count` INT NOT NULL DEFAULT 0,
  `collects_count` INT NOT NULL DEFAULT 0,
  `views_count` INT NOT NULL DEFAULT 0,
  `published_at` DATETIME,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_task_id` (`task_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_audit_status` (`audit_status`),
  KEY `idx_category_created` (`category`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='作品表';

-- 作品审核记录
DROP TABLE IF EXISTS `work_audits`;
CREATE TABLE `work_audits` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `work_id` BIGINT NOT NULL,
  `admin_id` BIGINT NOT NULL,
  `status` VARCHAR(20) NOT NULL COMMENT 'approved/rejected',
  `reason` VARCHAR(500),
  `audited_at` DATETIME NOT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_work_id` (`work_id`),
  KEY `idx_admin_id` (`admin_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='作品审核记录';

-- 作品点赞
DROP TABLE IF EXISTS `work_likes`;
CREATE TABLE `work_likes` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT NOT NULL,
  `work_id` BIGINT NOT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_work` (`user_id`, `work_id`),
  KEY `idx_work_id` (`work_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='作品点赞';

-- 作品收藏
DROP TABLE IF EXISTS `work_collects`;
CREATE TABLE `work_collects` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT NOT NULL,
  `work_id` BIGINT NOT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_work` (`user_id`, `work_id`),
  KEY `idx_work_id` (`work_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='作品收藏';

-- ----------------------------
-- 4. 积分模块
-- ----------------------------

-- 灵感值账户
DROP TABLE IF EXISTS `credit_accounts`;
CREATE TABLE `credit_accounts` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT NOT NULL,
  `balance` INT NOT NULL DEFAULT 0,
  `total_income` INT NOT NULL DEFAULT 0,
  `total_expense` INT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='灵感值账户';

-- 灵感值流水
DROP TABLE IF EXISTS `credit_ledgers`;
CREATE TABLE `credit_ledgers` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT NOT NULL,
  `type` VARCHAR(20) NOT NULL COMMENT 'income/expense',
  `amount` INT NOT NULL,
  `title` VARCHAR(255) NOT NULL,
  `source_type` VARCHAR(50) COMMENT 'recharge/task/checkin/redeem/gift/invite_register/invite_recharge',
  `source_id` BIGINT,
  `balance_after` INT NOT NULL,
  `idempotency_key` VARCHAR(100),
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_idempotency_key` (`idempotency_key`),
  KEY `idx_user_created` (`user_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='灵感值流水';

-- 签到记录
DROP TABLE IF EXISTS `checkin_records`;
CREATE TABLE `checkin_records` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT NOT NULL,
  `date` VARCHAR(10) NOT NULL COMMENT 'YYYY-MM-DD',
  `streak` INT NOT NULL DEFAULT 1 COMMENT '连续天数',
  `reward` INT NOT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_date` (`user_id`, `date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='签到记录';

-- 兑换码
DROP TABLE IF EXISTS `redeem_codes`;
CREATE TABLE `redeem_codes` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `code` VARCHAR(50) NOT NULL,
  `amount` INT NOT NULL,
  `batch_id` VARCHAR(50),
  `batch_name` VARCHAR(100),
  `used_by` BIGINT,
  `used_at` DATETIME,
  `expires_at` DATETIME,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_code` (`code`),
  KEY `idx_batch_id` (`batch_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='兑换码';

-- 充值套餐
DROP TABLE IF EXISTS `credit_packages`;
CREATE TABLE `credit_packages` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(50) NOT NULL,
  `price` DECIMAL(10,2) NOT NULL,
  `points` INT NOT NULL,
  `note` VARCHAR(255),
  `is_hot` TINYINT(1) NOT NULL DEFAULT 0,
  `sort_order` INT NOT NULL DEFAULT 0,
  `is_active` TINYINT(1) NOT NULL DEFAULT 1,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='充值套餐';

-- ----------------------------
-- 5. 邀请模块
-- ----------------------------

-- 邀请记录
DROP TABLE IF EXISTS `invitations`;
CREATE TABLE `invitations` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `inviter_id` BIGINT NOT NULL,
  `invitee_id` BIGINT NOT NULL,
  `invite_code` VARCHAR(10) NOT NULL,
  `status` VARCHAR(20) NOT NULL COMMENT 'registered/rewarded',
  `registered_at` DATETIME NOT NULL,
  `rewarded_at` DATETIME,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_inviter_id` (`inviter_id`),
  KEY `idx_invitee_id` (`invitee_id`),
  KEY `idx_invite_code` (`invite_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邀请记录';

-- 邀请奖励
DROP TABLE IF EXISTS `invitation_rewards`;
CREATE TABLE `invitation_rewards` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `invitation_id` BIGINT NOT NULL,
  `inviter_id` BIGINT NOT NULL,
  `invitee_id` BIGINT NOT NULL,
  `inviter_reward` INT NOT NULL,
  `invitee_reward` INT NOT NULL,
  `trigger_type` VARCHAR(20) NOT NULL COMMENT 'register/first_recharge',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_invitation_id` (`invitation_id`),
  KEY `idx_inviter_id` (`inviter_id`),
  KEY `idx_invitee_id` (`invitee_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邀请奖励';

-- ----------------------------
-- 6. 支付模块
-- ----------------------------

-- 支付订单
DROP TABLE IF EXISTS `payment_orders`;
CREATE TABLE `payment_orders` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT NOT NULL,
  `order_no` VARCHAR(50) NOT NULL,
  `package_id` BIGINT NOT NULL,
  `amount` DECIMAL(10,2) NOT NULL,
  `points` INT NOT NULL,
  `status` VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT 'pending/paid/failed/refunded',
  `channel_id` BIGINT NOT NULL,
  `pay_method_name` VARCHAR(50),
  `trade_no` VARCHAR(100),
  `paid_at` DATETIME,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_order_no` (`order_no`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_channel_id` (`channel_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付订单';

-- 支付交易记录
DROP TABLE IF EXISTS `payment_transactions`;
CREATE TABLE `payment_transactions` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `order_id` BIGINT NOT NULL,
  `channel_id` BIGINT NOT NULL,
  `transaction_id` VARCHAR(100),
  `amount` DECIMAL(10,2) NOT NULL,
  `status` VARCHAR(20) NOT NULL,
  `request_data` JSON,
  `callback_data` JSON,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_transaction_id` (`transaction_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付交易记录';

-- 支付渠道配置
DROP TABLE IF EXISTS `payment_channels`;
CREATE TABLE `payment_channels` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(50) NOT NULL,
  `channel_type` VARCHAR(20) NOT NULL COMMENT 'alipay/wechat/epay/demo',
  `merchant_id` VARCHAR(100),
  `config` JSON NOT NULL,
  `is_active` TINYINT(1) NOT NULL DEFAULT 1,
  `sort_order` INT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付渠道配置';

-- ----------------------------
-- 7. 配置模块
-- ----------------------------

-- AI模型配置
DROP TABLE IF EXISTS `ai_models`;
CREATE TABLE `ai_models` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(100) NOT NULL,
  `display_name` VARCHAR(100) NOT NULL,
  `type` VARCHAR(20) NOT NULL COMMENT 'image/video',
  `provider` VARCHAR(50) NOT NULL,
  `description` VARCHAR(500),
  `logo_url` VARCHAR(255),
  `tag` VARCHAR(50),
  `cost` INT NOT NULL,
  `api_config` JSON NOT NULL,
  `params_config` JSON,
  `is_active` TINYINT(1) NOT NULL DEFAULT 1,
  `sort_order` INT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_type_active` (`type`, `is_active`),
  KEY `idx_sort` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI模型配置';

-- 系统配置
DROP TABLE IF EXISTS `system_configs`;
CREATE TABLE `system_configs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `key` VARCHAR(100) NOT NULL,
  `value` TEXT NOT NULL,
  `type` VARCHAR(20) NOT NULL COMMENT 'string/int/json/bool',
  `category` VARCHAR(50),
  `description` VARCHAR(255),
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_key` (`key`),
  KEY `idx_category` (`category`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统配置';

-- 分类配置
DROP TABLE IF EXISTS `categories`;
CREATE TABLE `categories` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(50) NOT NULL,
  `type` VARCHAR(30) NOT NULL COMMENT 'work_category/aspect_ratio',
  `sort_order` INT NOT NULL DEFAULT 0,
  `is_active` TINYINT(1) NOT NULL DEFAULT 1,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='分类配置';

-- 邮箱配置
DROP TABLE IF EXISTS `email_config`;
CREATE TABLE `email_config` (
  `id` BIGINT NOT NULL DEFAULT 1,
  `smtp_host` VARCHAR(100) NOT NULL,
  `smtp_port` INT NOT NULL,
  `smtp_user` VARCHAR(100) NOT NULL,
  `smtp_password` VARCHAR(255) NOT NULL,
  `smtp_ssl` TINYINT(1) NOT NULL DEFAULT 0,
  `from_name` VARCHAR(100) NOT NULL,
  `from_email` VARCHAR(100) NOT NULL,
  `is_active` TINYINT(1) NOT NULL DEFAULT 1,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邮箱配置';

-- ----------------------------
-- 8. 管理后台模块
-- ----------------------------

-- 管理员
DROP TABLE IF EXISTS `admin_users`;
CREATE TABLE `admin_users` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `username` VARCHAR(50) NOT NULL,
  `password_hash` VARCHAR(255) NOT NULL,
  `name` VARCHAR(50) NOT NULL,
  `role` VARCHAR(20) NOT NULL COMMENT 'super_admin/admin/auditor',
  `permissions` JSON,
  `is_active` TINYINT(1) NOT NULL DEFAULT 1,
  `last_login_at` DATETIME,
  `last_login_ip` VARCHAR(50),
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理员';

-- 管理员操作日志
DROP TABLE IF EXISTS `admin_logs`;
CREATE TABLE `admin_logs` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `admin_id` BIGINT NOT NULL,
  `action` VARCHAR(50) NOT NULL,
  `target_type` VARCHAR(50),
  `target_id` BIGINT,
  `before_data` JSON,
  `after_data` JSON,
  `description` VARCHAR(500),
  `ip` VARCHAR(50),
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_admin_id` (`admin_id`),
  KEY `idx_action` (`action`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理员操作日志';

SET FOREIGN_KEY_CHECKS = 1;
