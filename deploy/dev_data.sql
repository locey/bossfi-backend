-- BossFi 开发环境测试数据脚本
-- 用于本地开发和测试环境的数据初始化

-- 确保连接到正确的数据库
\c bossfi;

-- 清理现有测试数据（仅开发环境使用）
TRUNCATE TABLE users RESTART IDENTITY CASCADE;

-- 插入开发测试用户数据
INSERT INTO users (
    wallet_address, 
    username, 
    email, 
    avatar, 
    bio, 
    boss_balance, 
    staked_amount, 
    reward_balance, 
    is_profile_complete,
    last_login_at,
    created_at,
    updated_at
) VALUES 
-- 1. 完整资料的活跃用户
(
    '0x1234567890123456789012345678901234567890',
    'alice_crypto',
    'alice@example.com',
    'https://avatars.githubusercontent.com/u/1?v=4',
    '区块链开发者，专注于DeFi和智能合约开发。热爱去中心化技术，致力于构建更好的Web3生态。',
    15000.50000000,
    8000.25000000,
    1250.75000000,
    true,
    NOW() - INTERVAL '2 hours',
    NOW() - INTERVAL '30 days',
    NOW() - INTERVAL '2 hours'
),

-- 2. 高净值用户
(
    '0x2345678901234567890123456789012345678901',
    'crypto_whale',
    'whale@example.com',
    'https://avatars.githubusercontent.com/u/2?v=4',
    '早期比特币投资者，现专注于DeFi生态系统投资。',
    100000.00000000,
    50000.00000000,
    5000.00000000,
    true,
    NOW() - INTERVAL '1 day',
    NOW() - INTERVAL '90 days',
    NOW() - INTERVAL '1 day'
),

-- 3. 新注册用户（资料未完善）
(
    '0x3456789012345678901234567890123456789012',
    NULL,
    NULL,
    NULL,
    NULL,
    0.00000000,
    0.00000000,
    0.00000000,
    false,
    NOW() - INTERVAL '10 minutes',
    NOW() - INTERVAL '10 minutes',
    NOW() - INTERVAL '10 minutes'
),

-- 4. 部分资料的用户
(
    '0x4567890123456789012345678901234567890123',
    'bob_defi',
    'bob@example.com',
    NULL,
    '刚开始接触DeFi，正在学习中...',
    500.12345678,
    200.00000000,
    25.50000000,
    false,
    NOW() - INTERVAL '5 hours',
    NOW() - INTERVAL '7 days',
    NOW() - INTERVAL '5 hours'
),

-- 5. 长期未登录用户
(
    '0x5678901234567890123456789012345678901234',
    'inactive_user',
    'inactive@example.com',
    'https://avatars.githubusercontent.com/u/5?v=4',
    '很久没有使用的账户',
    1000.00000000,
    500.00000000,
    100.00000000,
    true,
    NOW() - INTERVAL '60 days',
    NOW() - INTERVAL '120 days',
    NOW() - INTERVAL '60 days'
),

-- 6. 质押大户
(
    '0x6789012345678901234567890123456789012345',
    'staking_master',
    'staker@example.com',
    'https://avatars.githubusercontent.com/u/6?v=4',
    '专业的质押投资者，风险管理专家。',
    25000.00000000,
    20000.00000000,
    3000.00000000,
    true,
    NOW() - INTERVAL '6 hours',
    NOW() - INTERVAL '45 days',
    NOW() - INTERVAL '6 hours'
),

-- 7. 活跃交易用户
(
    '0x7890123456789012345678901234567890123456',
    'trader_pro',
    'trader@example.com',
    'https://avatars.githubusercontent.com/u/7?v=4',
    '全职交易员，专注于加密货币和DeFi协议交易。',
    8500.75000000,
    3000.25000000,
    850.12500000,
    true,
    NOW() - INTERVAL '30 minutes',
    NOW() - INTERVAL '15 days',
    NOW() - INTERVAL '30 minutes'
),

-- 8. 测试用户（用于自动化测试）
(
    '0x8901234567890123456789012345678901234567',
    'test_user',
    'test@bossfi.local',
    NULL,
    '这是一个测试账户，用于自动化测试。',
    100.00000000,
    50.00000000,
    10.00000000,
    true,
    NOW() - INTERVAL '1 hour',
    NOW() - INTERVAL '1 day',
    NOW() - INTERVAL '1 hour'
),

-- 9. 零余额用户
(
    '0x9012345678901234567890123456789012345678',
    'zero_balance',
    'zero@example.com',
    NULL,
    '刚注册的用户，还没有任何资产。',
    0.00000000,
    0.00000000,
    0.00000000,
    false,
    NOW() - INTERVAL '2 days',
    NOW() - INTERVAL '2 days',
    NOW() - INTERVAL '2 days'
),

-- 10. VIP用户
(
    '0xa123456789012345678901234567890123456789',
    'vip_investor',
    'vip@example.com',
    'https://avatars.githubusercontent.com/u/10?v=4',
    'BossFi VIP投资者，区块链行业资深人士，多个DeFi项目的早期投资人。',
    250000.00000000,
    150000.00000000,
    25000.00000000,
    true,
    NOW() - INTERVAL '15 minutes',
    NOW() - INTERVAL '180 days',
    NOW() - INTERVAL '15 minutes'
);

-- 插入一些额外的用户用于分页测试
INSERT INTO users (wallet_address, username, boss_balance, staked_amount, reward_balance, is_profile_complete, created_at, updated_at) 
SELECT 
    '0x' || lpad(to_hex(1000 + generate_series), 40, '0'),
    'user_' || generate_series,
    (random() * 10000)::decimal(20,8),
    (random() * 5000)::decimal(20,8),
    (random() * 1000)::decimal(20,8),
    random() > 0.3,
    NOW() - (random() * interval '365 days'),
    NOW() - (random() * interval '30 days')
FROM generate_series(1, 20);

-- 显示插入的数据统计
SELECT 
    COUNT(*) as total_users,
    COUNT(*) FILTER (WHERE is_profile_complete = true) as complete_profiles,
    COUNT(*) FILTER (WHERE boss_balance > 0) as users_with_balance,
    COUNT(*) FILTER (WHERE staked_amount > 0) as staking_users,
    ROUND(AVG(boss_balance), 2) as avg_balance,
    ROUND(SUM(boss_balance), 2) as total_balance
FROM users;

-- 显示最近登录的用户
SELECT 
    username,
    wallet_address,
    boss_balance,
    is_profile_complete,
    last_login_at
FROM users 
WHERE last_login_at IS NOT NULL 
ORDER BY last_login_at DESC 
LIMIT 5;

COMMIT;