-- AI评分功能数据库迁移脚本
-- 为articles表添加AI评分相关字段

-- 添加AI评分字段
ALTER TABLE articles 
ADD COLUMN IF NOT EXISTS score DECIMAL(3,2) DEFAULT NULL,
ADD COLUMN IF NOT EXISTS score_time TIMESTAMP DEFAULT NULL,
ADD COLUMN IF NOT EXISTS score_reason TEXT DEFAULT '',
ADD COLUMN IF NOT EXISTS score_status INT DEFAULT 0;

-- 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_articles_score ON articles(ai_score);
CREATE INDEX IF NOT EXISTS idx_articles_score_time ON articles(ai_score_time);
CREATE INDEX IF NOT EXISTS idx_articles_score_null ON articles(id) WHERE ai_score IS NULL;
CREATE INDEX IF NOT EXISTS idx_articles_score_status ON articles(score_status);
CREATE INDEX IF NOT EXISTS idx_articles_score_status_pending ON articles(id) WHERE score_status = 0;
CREATE INDEX IF NOT EXISTS idx_articles_score_status_failed ON articles(id) WHERE score_status = -1;


-- 添加注释
COMMENT ON COLUMN articles.score IS 'AI评分 (0-10分)';
COMMENT ON COLUMN articles.score_time IS 'AI评分时间';
COMMENT ON COLUMN articles.score_reason IS 'AI评分理由';
COMMENT ON COLUMN articles.score_status IS '评分状态: 0-待评分, 1-评分中, 2-评分成功, -1-评分失败';

-- 验证迁移
SELECT 
    column_name, 
    data_type, 
    is_nullable, 
    column_default
FROM information_schema.columns 
WHERE table_name = 'articles' 
AND column_name IN ('score', 'score_time', 'score_reason', 'score_status'); 