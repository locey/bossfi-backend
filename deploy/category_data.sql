-- BossFi 文章分类初始化数据

-- 插入默认分类
INSERT INTO article_categories (name, description, icon, color, sort_order, is_active) VALUES
('技术', '技术相关文章，包括编程、开发、架构等', 'tech-icon', '#FF5733', 1, true),
('区块链', '区块链技术、加密货币、DeFi等相关内容', 'blockchain-icon', '#33FF57', 2, true),
('投资', '投资理财、市场分析、投资策略等', 'investment-icon', '#3357FF', 3, true),
('生活', '日常生活、个人感悟、生活技巧等', 'life-icon', '#F3FF33', 4, true),
('新闻', '行业新闻、热点事件、重要公告等', 'news-icon', '#FF33F3', 5, true),
('教程', '学习教程、操作指南、最佳实践等', 'tutorial-icon', '#33FFF3', 6, true),
('观点', '个人观点、行业分析、深度思考等', 'opinion-icon', '#F333FF', 7, true),
('其他', '其他类型的内容', 'other-icon', '#999999', 8, true)
ON CONFLICT (name) DO NOTHING;

-- 更新序列（如果需要）
SELECT setval('article_categories_id_seq', (SELECT MAX(id) FROM article_categories)); 