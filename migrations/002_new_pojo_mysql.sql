-- =============================================================================
-- BossFi 数据库建表语句
-- 版本: 1.0.0
-- 创建时间: 2024-01-01
-- 描述: 区块链招聘论坛数据库表结构
-- =============================================================================

-- 设置字符集和外键检查
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- 创建数据库
CREATE DATABASE IF NOT EXISTS `bossfi` 
DEFAULT CHARACTER SET utf8mb4 
COLLATE utf8mb4_unicode_ci;

USE `bossfi`;

-- =============================================================================
-- 1. 用户表 (users)
-- 存储用户基本信息和钱包相关数据
-- =============================================================================
CREATE TABLE `users` (
  `id` char(36) NOT NULL COMMENT '用户ID (UUID)',
  `wallet_address` varchar(42) NOT NULL COMMENT '钱包地址',
  `username` varchar(50) DEFAULT NULL COMMENT '用户名',
  `email` varchar(255) DEFAULT NULL COMMENT '邮箱',
  `avatar` varchar(500) DEFAULT NULL COMMENT '头像URL',
  `bio` text COMMENT '个人简介',
  `boss_balance` decimal(36,18) DEFAULT '0.000000000000000000' COMMENT 'BOSS币余额',
  `staked_amount` decimal(36,18) DEFAULT '0.000000000000000000' COMMENT '质押金额',
  `reward_balance` decimal(36,18) DEFAULT '0.000000000000000000' COMMENT '奖励余额',
  `is_profile_complete` tinyint(1) DEFAULT '0' COMMENT '是否完善资料',
  `last_login_at` timestamp NULL DEFAULT NULL COMMENT '最后登录时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_wallet_address` (`wallet_address`),
  UNIQUE KEY `uk_username` (`username`),
  UNIQUE KEY `uk_email` (`email`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_last_login_at` (`last_login_at`),
  KEY `idx_boss_balance` (`boss_balance`),
  KEY `idx_staked_amount` (`staked_amount`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- =============================================================================
-- 2. 帖子表 (posts)  
-- 存储帖子信息，每个帖子对应一个NFT
-- =============================================================================
CREATE TABLE `posts` (
  `id` char(36) NOT NULL COMMENT '帖子ID (UUID)',
  `author_id` char(36) NOT NULL COMMENT '作者ID',
  `token_id` varchar(100) DEFAULT NULL COMMENT 'NFT Token ID',
  `title` varchar(255) NOT NULL COMMENT '标题',
  `content` longtext COMMENT '内容',
  `post_type` enum('job','resume','discussion') NOT NULL COMMENT '帖子类型',
  `status` enum('draft','published','closed') DEFAULT 'draft' COMMENT '状态',
  `tags` json DEFAULT NULL COMMENT '标签 (JSON数组)',
  `salary` varchar(100) DEFAULT NULL COMMENT '薪资',
  `location` varchar(100) DEFAULT NULL COMMENT '地点',
  `company` varchar(100) DEFAULT NULL COMMENT '公司',
  `requirements` text COMMENT '要求',
  `boss_cost` decimal(36,18) NOT NULL COMMENT '发帖消耗的BOSS币',
  `view_count` bigint DEFAULT '0' COMMENT '浏览次数',
  `like_count` bigint DEFAULT '0' COMMENT '点赞数',
  `reply_count` bigint DEFAULT '0' COMMENT '回复数',
  `ipfs_hash` varchar(255) DEFAULT NULL COMMENT 'IPFS哈希',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_token_id` (`token_id`),
  KEY `idx_author_id` (`author_id`),
  KEY `idx_post_type` (`post_type`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_view_count` (`view_count`),
  KEY `idx_like_count` (`like_count`),
  KEY `idx_boss_cost` (`boss_cost`),
  
  CONSTRAINT `fk_posts_author` FOREIGN KEY (`author_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='帖子表';

-- =============================================================================
-- 3. 质押表 (stakes)
-- 存储用户质押记录和状态
-- =============================================================================
CREATE TABLE `stakes` (
  `id` char(36) NOT NULL COMMENT '质押ID (UUID)',
  `user_id` char(36) NOT NULL COMMENT '用户ID',
  `amount` decimal(36,18) NOT NULL COMMENT '质押金额',
  `reward_earned` decimal(36,18) DEFAULT '0.000000000000000000' COMMENT '已获得奖励',
  `status` enum('active','unstaking','completed') DEFAULT 'active' COMMENT '状态',
  `staked_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '质押时间',
  `unstake_request_at` timestamp NULL DEFAULT NULL COMMENT '解质押请求时间',
  `unstaked_at` timestamp NULL DEFAULT NULL COMMENT '解质押完成时间',
  `last_reward_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '最后奖励时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_staked_at` (`staked_at`),
  KEY `idx_unstake_request_at` (`unstake_request_at`),
  KEY `idx_amount` (`amount`),
  
  CONSTRAINT `fk_stakes_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='质押表';

-- =============================================================================
-- 4. 奖励历史表 (reward_histories)
-- 记录所有奖励发放历史
-- =============================================================================
CREATE TABLE `reward_histories` (
  `id` char(36) NOT NULL COMMENT '奖励ID (UUID)',
  `user_id` char(36) NOT NULL COMMENT '用户ID',
  `stake_id` char(36) NOT NULL COMMENT '质押ID',
  `amount` decimal(36,18) NOT NULL COMMENT '奖励金额',
  `block_hash` varchar(66) DEFAULT NULL COMMENT '区块哈希',
  `tx_hash` varchar(66) DEFAULT NULL COMMENT '交易哈希',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_stake_id` (`stake_id`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_tx_hash` (`tx_hash`),
  KEY `idx_amount` (`amount`),
  
  CONSTRAINT `fk_reward_histories_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_reward_histories_stake` FOREIGN KEY (`stake_id`) REFERENCES `stakes` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='奖励历史表';

-- =============================================================================
-- 5. 帖子点赞表 (post_likes)
-- 记录用户对帖子的点赞
-- =============================================================================
CREATE TABLE `post_likes` (
  `id` char(36) NOT NULL COMMENT 'ID (UUID)',
  `post_id` char(36) NOT NULL COMMENT '帖子ID',
  `user_id` char(36) NOT NULL COMMENT '用户ID',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_post_user` (`post_id`,`user_id`),
  KEY `idx_post_id` (`post_id`),
  KEY `idx_user_id` (`user_id`),
  
  CONSTRAINT `fk_post_likes_post` FOREIGN KEY (`post_id`) REFERENCES `posts` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_post_likes_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='帖子点赞表';

-- =============================================================================
-- 6. 帖子回复表 (post_replies)
-- 存储帖子的回复和评论
-- =============================================================================
CREATE TABLE `post_replies` (
  `id` char(36) NOT NULL COMMENT '回复ID (UUID)',
  `post_id` char(36) NOT NULL COMMENT '帖子ID',
  `user_id` char(36) NOT NULL COMMENT '用户ID',
  `parent_id` char(36) DEFAULT NULL COMMENT '父回复ID',
  `content` text NOT NULL COMMENT '回复内容',
  `like_count` bigint DEFAULT '0' COMMENT '点赞数',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (`id`),
  KEY `idx_post_id` (`post_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_created_at` (`created_at`),
  
  CONSTRAINT `fk_post_replies_post` FOREIGN KEY (`post_id`) REFERENCES `posts` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_post_replies_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_post_replies_parent` FOREIGN KEY (`parent_id`) REFERENCES `post_replies` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='帖子回复表';

-- =============================================================================
-- 7. 系统配置表 (system_configs)
-- 存储系统配置参数
-- =============================================================================
CREATE TABLE `system_configs` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `config_key` varchar(100) NOT NULL COMMENT '配置键',
  `config_value` text COMMENT '配置值',
  `description` varchar(255) DEFAULT NULL COMMENT '描述',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_config_key` (`config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';

-- =============================================================================
-- 8. 用户会话表 (user_sessions)
-- 存储用户登录会话信息
-- =============================================================================
CREATE TABLE `user_sessions` (
  `id` char(36) NOT NULL COMMENT '会话ID (UUID)',
  `user_id` char(36) NOT NULL COMMENT '用户ID',
  `wallet_address` varchar(42) NOT NULL COMMENT '钱包地址',
  `nonce` varchar(255) NOT NULL COMMENT '登录随机数',
  `token` varchar(500) NOT NULL COMMENT 'JWT Token',
  `expires_at` timestamp NOT NULL COMMENT '过期时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_wallet_address` (`wallet_address`),
  KEY `idx_expires_at` (`expires_at`),
  KEY `idx_token` (`token`(255)),
  
  CONSTRAINT `fk_user_sessions_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户会话表';

-- =============================================================================
-- 初始化系统配置数据
-- =============================================================================
INSERT INTO `system_configs` (`config_key`, `config_value`, `description`) VALUES
('boss_token_address', '', 'BOSS代币合约地址'),
('nft_contract_address', '', 'NFT合约地址'),
('staking_contract_address', '', '质押合约地址'),
('post_cost_boss', '10', '发帖消耗BOSS币数量'),
('annual_reward_rate', '0.10', '年化奖励率'),
('unstake_delay_days', '7', '解质押延迟天数'),
('min_stake_amount', '0.001', '最小质押金额(ETH)'),
('eth_to_boss_rate', '1000', 'ETH到BOSS的兑换率'),
('max_post_title_length', '255', '帖子标题最大长度'),
('max_post_content_length', '10000', '帖子内容最大长度'),
('max_reply_length', '1000', '回复最大长度'),
('platform_fee_rate', '0.05', '平台手续费率');

-- =============================================================================
-- 创建有用的视图
-- =============================================================================

-- 用户统计视图
CREATE OR REPLACE VIEW `user_stats` AS
SELECT 
    u.id,
    u.wallet_address,
    u.username,
    u.email,
    u.boss_balance,
    u.staked_amount,
    u.reward_balance,
    u.is_profile_complete,
    COALESCE(p.post_count, 0) as post_count,
    COALESCE(p.published_posts, 0) as published_posts,
    COALESCE(s.active_stakes, 0) as active_stakes,
    COALESCE(s.total_staked, 0) as total_staked,
    COALESCE(r.total_rewards, 0) as total_rewards,
    COALESCE(l.total_likes_received, 0) as total_likes_received,
    u.created_at,
    u.last_login_at
FROM users u
LEFT JOIN (
    SELECT 
        author_id, 
        COUNT(*) as post_count,
        SUM(CASE WHEN status = 'published' THEN 1 ELSE 0 END) as published_posts
    FROM posts 
    GROUP BY author_id
) p ON u.id = p.author_id
LEFT JOIN (
    SELECT 
        user_id, 
        COUNT(*) as active_stakes,
        SUM(amount) as total_staked
    FROM stakes 
    WHERE status = 'active' 
    GROUP BY user_id
) s ON u.id = s.user_id
LEFT JOIN (
    SELECT 
        user_id, 
        SUM(amount) as total_rewards 
    FROM reward_histories 
    GROUP BY user_id
) r ON u.id = r.user_id
LEFT JOIN (
    SELECT 
        p.author_id,
        SUM(p.like_count) as total_likes_received
    FROM posts p
    GROUP BY p.author_id
) l ON u.id = l.author_id;

-- 帖子详情视图
CREATE OR REPLACE VIEW `post_details` AS
SELECT 
    p.id,
    p.token_id,
    p.title,
    p.content,
    p.post_type,
    p.status,
    p.tags,
    p.salary,
    p.location,
    p.company,
    p.requirements,
    p.boss_cost,
    p.view_count,
    p.like_count,
    p.reply_count,
    p.ipfs_hash,
    u.username as author_name,
    u.wallet_address as author_address,
    u.avatar as author_avatar,
    p.created_at,
    p.updated_at
FROM posts p
LEFT JOIN users u ON p.author_id = u.id;

-- 质押统计视图
CREATE OR REPLACE VIEW `staking_stats` AS
SELECT 
    DATE(created_at) as stake_date,
    COUNT(*) as daily_stakes,
    SUM(amount) as daily_amount,
    AVG(amount) as avg_amount,
    COUNT(CASE WHEN status = 'active' THEN 1 END) as active_stakes,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_stakes
FROM stakes
GROUP BY DATE(created_at)
ORDER BY stake_date DESC;

-- 恢复外键检查
SET FOREIGN_KEY_CHECKS = 1;

-- =============================================================================
-- 建表完成
-- =============================================================================
SELECT 'BossFi 数据库表结构创建完成！' as message,
       'Total Tables: 8' as table_count,
       'Total Views: 3' as view_count; 