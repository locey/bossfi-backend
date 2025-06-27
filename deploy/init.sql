-- BossFi 数据库初始化脚本

-- 创建数据库（如果不存在）
SELECT 'CREATE DATABASE bossfi'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'bossfi')\gexec

-- 连接到 bossfi 数据库
\c bossfi;

-- 创建用户表（如果不存在）
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    wallet_address VARCHAR(42) UNIQUE NOT NULL,
    username VARCHAR(100),
    email VARCHAR(255),
    avatar TEXT,
    bio TEXT,
    boss_balance DECIMAL(20, 8) DEFAULT 0,
    staked_amount DECIMAL(20, 8) DEFAULT 0,
    reward_balance DECIMAL(20, 8) DEFAULT 0,
    is_profile_complete BOOLEAN DEFAULT FALSE,
    nonce VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_wallet_address ON users(wallet_address);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
CREATE INDEX IF NOT EXISTS idx_users_last_login_at ON users(last_login_at);

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 创建触发器
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 插入测试数据（仅在开发环境）
-- INSERT INTO users (wallet_address, username, boss_balance, staked_amount, reward_balance, is_profile_complete)
-- VALUES 
--     ('0x1234567890123456789012345678901234567890', 'test_user_1', 1000.0, 500.0, 50.0, true),
--     ('0x0987654321098765432109876543210987654321', 'test_user_2', 2000.0, 1000.0, 100.0, false)
-- ON CONFLICT (wallet_address) DO NOTHING;

-- 权限设置
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO bossfi_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO bossfi_user; 