-- =============================================================================
-- BossFi PostgreSQL 数据库建表语句
-- 版本: 1.0.0
-- 创建时间: 2024-01-01  
-- 描述: 区块链招聘论坛数据库表结构 (PostgreSQL)
-- =============================================================================

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================================================
-- 1. 用户表 (users)
-- 存储用户基本信息和钱包相关数据
-- =============================================================================
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  wallet_address VARCHAR(42) NOT NULL UNIQUE,
  username VARCHAR(50) UNIQUE,
  email VARCHAR(255) UNIQUE,
  avatar VARCHAR(500),
  bio TEXT,
  boss_balance DECIMAL(36,18) DEFAULT 0,
  staked_amount DECIMAL(36,18) DEFAULT 0,
  reward_balance DECIMAL(36,18) DEFAULT 0,
  is_profile_complete BOOLEAN DEFAULT FALSE,
  last_login_at TIMESTAMP WITH TIME ZONE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_users_last_login_at ON users(last_login_at);
CREATE INDEX idx_users_boss_balance ON users(boss_balance);
CREATE INDEX idx_users_staked_amount ON users(staked_amount);

-- 创建更新时间触发器
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users 
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- 2. 帖子表 (posts)
-- 存储帖子信息，每个帖子对应一个NFT
-- =============================================================================
CREATE TYPE post_type_enum AS ENUM ('job', 'resume', 'discussion');
CREATE TYPE post_status_enum AS ENUM ('draft', 'published', 'closed');

CREATE TABLE posts (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_id VARCHAR(100) UNIQUE,
  title VARCHAR(255) NOT NULL,
  content TEXT,
  post_type post_type_enum NOT NULL,
  status post_status_enum DEFAULT 'draft',
  tags JSONB,
  salary VARCHAR(100),
  location VARCHAR(100),
  company VARCHAR(100),
  requirements TEXT,
  boss_cost DECIMAL(36,18) NOT NULL,
  view_count BIGINT DEFAULT 0,
  like_count BIGINT DEFAULT 0,
  reply_count BIGINT DEFAULT 0,
  ipfs_hash VARCHAR(255),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_posts_author_id ON posts(author_id);
CREATE INDEX idx_posts_post_type ON posts(post_type);
CREATE INDEX idx_posts_status ON posts(status);
CREATE INDEX idx_posts_created_at ON posts(created_at);
CREATE INDEX idx_posts_view_count ON posts(view_count);
CREATE INDEX idx_posts_like_count ON posts(like_count);
CREATE INDEX idx_posts_boss_cost ON posts(boss_cost);

-- 创建更新时间触发器
CREATE TRIGGER update_posts_updated_at BEFORE UPDATE ON posts 
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- 3. 质押表 (stakes)
-- 存储用户质押记录和状态
-- =============================================================================
CREATE TYPE stake_status_enum AS ENUM ('active', 'unstaking', 'completed');

CREATE TABLE stakes (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  amount DECIMAL(36,18) NOT NULL,
  reward_earned DECIMAL(36,18) DEFAULT 0,
  status stake_status_enum DEFAULT 'active',
  staked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  unstake_request_at TIMESTAMP WITH TIME ZONE,
  unstaked_at TIMESTAMP WITH TIME ZONE,
  last_reward_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_stakes_user_id ON stakes(user_id);
CREATE INDEX idx_stakes_status ON stakes(status);
CREATE INDEX idx_stakes_staked_at ON stakes(staked_at);
CREATE INDEX idx_stakes_unstake_request_at ON stakes(unstake_request_at);
CREATE INDEX idx_stakes_amount ON stakes(amount);

-- 创建更新时间触发器
CREATE TRIGGER update_stakes_updated_at BEFORE UPDATE ON stakes 
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- 4. 奖励历史表 (reward_histories)
-- 记录所有奖励发放历史
-- =============================================================================
CREATE TABLE reward_histories (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  stake_id UUID NOT NULL REFERENCES stakes(id) ON DELETE CASCADE,
  amount DECIMAL(36,18) NOT NULL,
  block_hash VARCHAR(66),
  tx_hash VARCHAR(66),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_reward_histories_user_id ON reward_histories(user_id);
CREATE INDEX idx_reward_histories_stake_id ON reward_histories(stake_id);
CREATE INDEX idx_reward_histories_created_at ON reward_histories(created_at);
CREATE INDEX idx_reward_histories_tx_hash ON reward_histories(tx_hash);
CREATE INDEX idx_reward_histories_amount ON reward_histories(amount);

-- =============================================================================
-- 5. 帖子点赞表 (post_likes)
-- 记录用户对帖子的点赞
-- =============================================================================
CREATE TABLE post_likes (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(post_id, user_id)
);

-- 创建索引
CREATE INDEX idx_post_likes_post_id ON post_likes(post_id);
CREATE INDEX idx_post_likes_user_id ON post_likes(user_id);

-- =============================================================================
-- 6. 帖子回复表 (post_replies)
-- 存储帖子的回复和评论
-- =============================================================================
CREATE TABLE post_replies (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  parent_id UUID REFERENCES post_replies(id) ON DELETE CASCADE,
  content TEXT NOT NULL,
  like_count BIGINT DEFAULT 0,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_post_replies_post_id ON post_replies(post_id);
CREATE INDEX idx_post_replies_user_id ON post_replies(user_id);
CREATE INDEX idx_post_replies_parent_id ON post_replies(parent_id);
CREATE INDEX idx_post_replies_created_at ON post_replies(created_at);

-- 创建更新时间触发器
CREATE TRIGGER update_post_replies_updated_at BEFORE UPDATE ON post_replies 
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- 7. 系统配置表 (system_configs)
-- 存储系统配置参数
-- =============================================================================
CREATE TABLE system_configs (
  id SERIAL PRIMARY KEY,
  config_key VARCHAR(100) NOT NULL UNIQUE,
  config_value TEXT,
  description VARCHAR(255),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建更新时间触发器
CREATE TRIGGER update_system_configs_updated_at BEFORE UPDATE ON system_configs 
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- 8. 用户会话表 (user_sessions)
-- 存储用户登录会话信息
-- =============================================================================
CREATE TABLE user_sessions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash VARCHAR(64) NOT NULL UNIQUE,
  refresh_token_hash VARCHAR(64),
  device_info TEXT,
  ip_address INET,
  user_agent TEXT,
  expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_token_hash ON user_sessions(token_hash);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);

-- 创建更新时间触发器
CREATE TRIGGER update_user_sessions_updated_at BEFORE UPDATE ON user_sessions 
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- 9. 视图 - 用户统计
-- =============================================================================
CREATE VIEW user_stats AS
SELECT 
    u.id,
    u.wallet_address,
    u.username,
    u.boss_balance,
    u.staked_amount,
    u.reward_balance,
    COALESCE(post_stats.post_count, 0) as post_count,
    COALESCE(post_stats.total_likes, 0) as total_likes,
    COALESCE(post_stats.total_views, 0) as total_views,
    COALESCE(stake_stats.total_rewards, 0) as total_rewards,
    u.created_at
FROM users u
LEFT JOIN (
    SELECT 
        author_id,
        COUNT(*) as post_count,
        SUM(like_count) as total_likes,
        SUM(view_count) as total_views
    FROM posts 
    WHERE status = 'published'
    GROUP BY author_id
) post_stats ON u.id = post_stats.author_id
LEFT JOIN (
    SELECT 
        user_id,
        SUM(reward_earned) as total_rewards
    FROM stakes 
    GROUP BY user_id
) stake_stats ON u.id = stake_stats.user_id;

-- =============================================================================
-- 10. 视图 - 帖子详情
-- =============================================================================
CREATE VIEW post_details AS
SELECT 
    p.*,
    u.username as author_username,
    u.wallet_address as author_wallet_address,
    u.avatar as author_avatar
FROM posts p
JOIN users u ON p.author_id = u.id;

-- =============================================================================
-- 11. 视图 - 质押统计
-- =============================================================================
CREATE VIEW staking_stats AS
SELECT 
    COUNT(*) as total_stakers,
    SUM(amount) as total_staked,
    SUM(reward_earned) as total_rewards_distributed,
    AVG(amount) as avg_stake_amount
FROM stakes 
WHERE status = 'active';

-- =============================================================================
-- 12. 初始化数据
-- =============================================================================

-- 插入系统配置
INSERT INTO system_configs (config_key, config_value, description) VALUES
('post_boss_cost', '10.0', '发帖消耗的BOSS币数量'),
('staking_apr', '12.0', '质押年化收益率 (%)'),
('min_stake_amount', '100.0', '最小质押金额'),
('unstaking_period', '7', '解质押等待期 (天)'),
('reward_interval', '24', '奖励发放间隔 (小时)');

-- 创建管理员用户 (测试用)
INSERT INTO users (id, wallet_address, username, email, boss_balance, is_profile_complete) VALUES
(uuid_generate_v4(), '0x1234567890123456789012345678901234567890', 'admin', 'admin@bossfi.io', 10000.0, true);

-- =============================================================================
-- 完成
-- =============================================================================
COMMENT ON DATABASE bossfi IS 'BossFi区块链招聘论坛数据库'; 