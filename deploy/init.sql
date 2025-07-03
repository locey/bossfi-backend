-- BossFi 数据库初始化脚本

-- 创建数据库（如果不存在）
SELECT 'CREATE DATABASE bossfi'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'bossfi')\gexec

-- 连接到 bossfi 数据库
\c bossfi;

-- =============================================================================
-- 1. 用户表 (users)
-- =============================================================================
CREATE TABLE IF NOT EXISTS users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
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

-- 表注释
COMMENT ON TABLE users IS '用户表 - 存储平台用户的基本信息、钱包地址和资产数据';
-- 字段注释
COMMENT ON COLUMN users.id IS '用户ID - 系统自动生成的唯一标识符';
COMMENT ON COLUMN users.wallet_address IS '钱包地址 - 用户的以太坊钱包地址，40位十六进制字符串';
COMMENT ON COLUMN users.username IS '用户名 - 用户在平台上的显示名称';
COMMENT ON COLUMN users.email IS '邮箱 - 用户邮箱地址，用于通知和账户找回';
COMMENT ON COLUMN users.avatar IS '头像 - 用户头像图片的URL链接';
COMMENT ON COLUMN users.bio IS '个人简介 - 用户的自我介绍和背景描述';
COMMENT ON COLUMN users.boss_balance IS 'BOSS代币余额 - 用户持有的BOSS代币数量';
COMMENT ON COLUMN users.staked_amount IS '质押数量 - 用户质押挖矿的代币数量';
COMMENT ON COLUMN users.reward_balance IS '奖励余额 - 用户通过质押等方式获得的奖励代币';
COMMENT ON COLUMN users.is_profile_complete IS '资料完整性 - 标记用户是否完成了基本资料填写';
COMMENT ON COLUMN users.nonce IS '随机数 - 用于钱包签名验证，防止重放攻击';
COMMENT ON COLUMN users.created_at IS '创建时间 - 用户注册时间';
COMMENT ON COLUMN users.updated_at IS '更新时间 - 用户信息最后更新时间';
COMMENT ON COLUMN users.last_login_at IS '最后登录时间 - 用户最近一次登录的时间';

-- =============================================================================
-- 2. 文章分类表 (article_categories)
-- =============================================================================
CREATE TABLE IF NOT EXISTS article_categories (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description VARCHAR(200),
    icon VARCHAR(100),
    color VARCHAR(7), -- 十六进制颜色值，如 #FF5733
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_article_categories_sort_order ON article_categories(sort_order);
CREATE INDEX IF NOT EXISTS idx_article_categories_is_active ON article_categories(is_active);

-- article_categories表注释
COMMENT ON TABLE article_categories IS '文章分类表 - 存储文章分类信息';
COMMENT ON COLUMN article_categories.id IS '分类ID - 系统自动生成的唯一标识符';
COMMENT ON COLUMN article_categories.name IS '分类名称 - 分类的显示名称，唯一';
COMMENT ON COLUMN article_categories.description IS '分类描述 - 分类的详细描述';
COMMENT ON COLUMN article_categories.icon IS '分类图标 - 分类的图标标识';
COMMENT ON COLUMN article_categories.color IS '分类颜色 - 分类的显示颜色，十六进制格式';
COMMENT ON COLUMN article_categories.sort_order IS '排序顺序 - 分类的显示排序，数字越小越靠前';
COMMENT ON COLUMN article_categories.is_active IS '是否活跃 - 标记分类是否可用';
COMMENT ON COLUMN article_categories.created_at IS '创建时间 - 分类创建时间';
COMMENT ON COLUMN article_categories.updated_at IS '更新时间 - 分类最后更新时间';

-- =============================================================================
-- 3. 文章表 (articles)
-- =============================================================================
CREATE TABLE IF NOT EXISTS articles (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL,
    category_id BIGINT, -- 分类ID，可为空
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    images JSONB,
    like_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    view_count INTEGER DEFAULT 0,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_articles_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_articles_category_id FOREIGN KEY (category_id) REFERENCES article_categories(id) ON DELETE SET NULL
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_articles_user_id ON articles(user_id);
CREATE INDEX IF NOT EXISTS idx_articles_created_at ON articles(created_at);

-- article表注释
COMMENT ON TABLE articles IS '文章表 - 存储用户发布的文章信息';
COMMENT ON COLUMN articles.id IS '文章ID - 系统自动生成的唯一标识符';
COMMENT ON COLUMN articles.user_id IS '用户ID - 关联到users表的用户ID';
COMMENT ON COLUMN articles.title IS '文章标题 - 文章的标题';
COMMENT ON COLUMN articles.content IS '文章内容 - 文章的正文内容';
COMMENT ON COLUMN articles.images IS '图片列表 - 文章的图片列表';
COMMENT ON COLUMN articles.like_count IS '点赞数 - 文章的点赞数量';
COMMENT ON COLUMN articles.comment_count IS '评论数 - 文章的评论数量';
COMMENT ON COLUMN articles.view_count IS '浏览数 - 文章的浏览数量';
COMMENT ON COLUMN articles.is_deleted IS '是否删除 - 标记文章是否被删除';
COMMENT ON COLUMN articles.created_at IS '创建时间 - 文章创建时间';
COMMENT ON COLUMN articles.updated_at IS '更新时间 - 文章最后更新时间';

-- =============================================================================
-- 3. 文章点赞表 (article_likes)
-- =============================================================================
CREATE TABLE IF NOT EXISTS article_likes (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL,
    article_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_article_likes_unique UNIQUE (article_id, user_id),
    CONSTRAINT fk_article_likes_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_article_likes_article_id FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_article_likes_user_id ON article_likes(user_id);
CREATE INDEX IF NOT EXISTS idx_article_likes_article_id ON article_likes(article_id);
CREATE INDEX IF NOT EXISTS idx_article_likes_created_at ON article_likes(created_at);

-- article_likes表注释
COMMENT ON TABLE article_likes IS '文章点赞表 - 存储用户对文章的点赞信息';
COMMENT ON COLUMN article_likes.id IS '点赞ID - 系统自动生成的唯一标识符';
COMMENT ON COLUMN article_likes.user_id IS '用户ID - 关联到users表的用户ID';
COMMENT ON COLUMN article_likes.article_id IS '文章ID - 关联到articles表的文章ID';
COMMENT ON COLUMN article_likes.created_at IS '创建时间 - 点赞创建时间';

-- =============================================================================
-- 4. 文章评论表 (article_comments)
-- =============================================================================
CREATE TABLE IF NOT EXISTS article_comments (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL,
    article_id BIGINT NOT NULL,
    parent_id BIGINT, -- 父评论ID，用于回复功能
    content TEXT NOT NULL,
    like_count INTEGER DEFAULT 0,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_article_comments_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_article_comments_article_id FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE,
    CONSTRAINT fk_article_comments_parent_id FOREIGN KEY (parent_id) REFERENCES article_comments(id) ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_article_comments_user_id ON article_comments(user_id);
CREATE INDEX IF NOT EXISTS idx_article_comments_article_id ON article_comments(article_id);
CREATE INDEX IF NOT EXISTS idx_article_comments_parent_id ON article_comments(parent_id);
CREATE INDEX IF NOT EXISTS idx_article_comments_created_at ON article_comments(created_at);

-- article_comments表注释
COMMENT ON TABLE article_comments IS '文章评论表 - 存储用户对文章的评论信息';
COMMENT ON COLUMN article_comments.id IS '评论ID - 系统自动生成的唯一标识符';
COMMENT ON COLUMN article_comments.user_id IS '用户ID - 关联到users表的用户ID';
COMMENT ON COLUMN article_comments.article_id IS '文章ID - 关联到articles表的文章ID';
COMMENT ON COLUMN article_comments.parent_id IS '父评论ID - 用于回复功能';
COMMENT ON COLUMN article_comments.content IS '评论内容 - 用户对文章的评论内容';
COMMENT ON COLUMN article_comments.like_count IS '点赞数 - 评论的点赞数量';
COMMENT ON COLUMN article_comments.is_deleted IS '是否删除 - 标记评论是否被删除';
COMMENT ON COLUMN article_comments.created_at IS '创建时间 - 评论创建时间';

-- =============================================================================
-- 5. 文章评论点赞表 (article_comment_likes)
-- =============================================================================
CREATE TABLE IF NOT EXISTS article_comment_likes (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL,
    comment_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_article_comment_likes_unique UNIQUE (comment_id, user_id),
    CONSTRAINT fk_article_comment_likes_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_article_comment_likes_comment_id FOREIGN KEY (comment_id) REFERENCES article_comments(id) ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_article_comment_likes_user_id ON article_comment_likes(user_id);
CREATE INDEX IF NOT EXISTS idx_article_comment_likes_comment_id ON article_comment_likes(comment_id);
CREATE INDEX IF NOT EXISTS idx_article_comment_likes_created_at ON article_comment_likes(created_at);

-- article_comment_likes表注释 
COMMENT ON TABLE article_comment_likes IS '文章评论点赞表 - 存储用户对文章评论的点赞信息';
COMMENT ON COLUMN article_comment_likes.id IS '点赞ID - 系统自动生成的唯一标识符';
COMMENT ON COLUMN article_comment_likes.user_id IS '用户ID - 关联到users表的用户ID';
COMMENT ON COLUMN article_comment_likes.comment_id IS '评论ID - 关联到article_comments表的评论ID';
COMMENT ON COLUMN article_comment_likes.created_at IS '创建时间 - 点赞创建时间';