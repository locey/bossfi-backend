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