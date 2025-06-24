-- =============================================================================
-- PostgreSQL 权限修复脚本
-- 解决bossfier用户看不到表的问题
-- =============================================================================

-- 以postgres用户身份连接执行此脚本
-- psql -U postgres -h localhost -p 5432 -d bossfi -f fix_permissions.sql

-- 1. 检查当前状态
\echo '=== 修复前状态检查 ===';
SELECT current_user as executing_as;
SELECT COUNT(*) as total_tables FROM pg_tables WHERE schemaname = 'public';

-- 2. 确保bossfier用户存在并设置权限
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'bossfier') THEN
        CREATE USER bossfier WITH PASSWORD 'bossfier' CREATEDB LOGIN;
        RAISE NOTICE 'Created user bossfier';
    END IF;
END $$;

-- 3. 授予数据库级权限
GRANT ALL PRIVILEGES ON DATABASE bossfi TO bossfier;
ALTER DATABASE bossfi OWNER TO bossfier;

-- 4. 授予schema权限
GRANT USAGE ON SCHEMA public TO bossfier;
GRANT CREATE ON SCHEMA public TO bossfier;
GRANT ALL ON SCHEMA public TO bossfier;

-- 5. 授予所有现有表的权限
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO bossfier;

-- 6. 授予所有现有序列的权限  
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO bossfier;

-- 7. 授予所有现有函数的权限
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO bossfier;

-- 8. 设置默认权限（对未来创建的对象）
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO bossfier;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO bossfier;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON FUNCTIONS TO bossfier;

-- 9. 如果表的所有者不是bossfier，将所有权转移给bossfier
DO $$
DECLARE
    rec RECORD;
BEGIN
    FOR rec IN SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tableowner != 'bossfier'
    LOOP
        EXECUTE 'ALTER TABLE public.' || quote_ident(rec.tablename) || ' OWNER TO bossfier';
        RAISE NOTICE 'Changed owner of table % to bossfier', rec.tablename;
    END LOOP;
END $$;

-- 10. 同样处理序列
DO $$
DECLARE
    rec RECORD;
BEGIN
    FOR rec IN SELECT sequencename FROM pg_sequences WHERE schemaname = 'public'
    LOOP
        EXECUTE 'ALTER SEQUENCE public.' || quote_ident(rec.sequencename) || ' OWNER TO bossfier';
        RAISE NOTICE 'Changed owner of sequence % to bossfier', rec.sequencename;
    END LOOP;
END $$;

-- 11. 验证修复结果
\echo '=== 修复后状态检查 ===';

-- 显示所有表和所有者
SELECT 
    tablename,
    tableowner
FROM pg_tables 
WHERE schemaname = 'public'
ORDER BY tablename;

-- 检查bossfier用户权限
SELECT 
    'Database privileges' as check_type,
    has_database_privilege('bossfier', 'bossfi', 'CONNECT') as connect,
    has_database_privilege('bossfier', 'bossfi', 'CREATE') as create;

SELECT 
    'Schema privileges' as check_type,
    has_schema_privilege('bossfier', 'public', 'USAGE') as usage,
    has_schema_privilege('bossfier', 'public', 'CREATE') as create;

-- 检查表权限
SELECT 
    tablename,
    has_table_privilege('bossfier', 'public.' || tablename, 'SELECT') as can_select,
    has_table_privilege('bossfier', 'public.' || tablename, 'INSERT') as can_insert
FROM pg_tables 
WHERE schemaname = 'public'
ORDER BY tablename;

\echo '=== 权限修复完成 ===';
\echo '现在用bossfier用户连接应该能看到所有表了';
\echo '连接命令: psql -U bossfier -h localhost -p 5432 -d bossfi'; 