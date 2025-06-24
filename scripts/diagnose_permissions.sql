-- =============================================================================
-- PostgreSQL 权限诊断脚本
-- 用于检查用户权限和表访问问题
-- =============================================================================

-- 1. 检查当前连接信息
SELECT current_user as current_user, current_database() as current_database;

-- 2. 检查数据库中的所有表
SELECT 
    schemaname,
    tablename,
    tableowner
FROM pg_tables 
WHERE schemaname = 'public'
ORDER BY tablename;

-- 3. 检查用户对数据库的权限
SELECT 
    datname,
    has_database_privilege(current_user, datname, 'CONNECT') as can_connect,
    has_database_privilege(current_user, datname, 'CREATE') as can_create
FROM pg_database 
WHERE datname = 'bossfi';

-- 4. 检查用户对schema的权限
SELECT 
    nspname as schema_name,
    has_schema_privilege(current_user, nspname, 'USAGE') as has_usage,
    has_schema_privilege(current_user, nspname, 'CREATE') as has_create
FROM pg_namespace 
WHERE nspname = 'public';

-- 5. 检查用户对表的权限
SELECT 
    schemaname,
    tablename,
    has_table_privilege(current_user, schemaname||'.'||tablename, 'SELECT') as can_select,
    has_table_privilege(current_user, schemaname||'.'||tablename, 'INSERT') as can_insert,
    has_table_privilege(current_user, schemaname||'.'||tablename, 'UPDATE') as can_update,
    has_table_privilege(current_user, schemaname||'.'||tablename, 'DELETE') as can_delete
FROM pg_tables 
WHERE schemaname = 'public';

-- 6. 检查表权限详情
SELECT 
    grantee,
    table_schema,
    table_name,
    privilege_type
FROM information_schema.table_privileges 
WHERE table_schema = 'public'
AND grantee = current_user
ORDER BY table_name, privilege_type;

-- 7. 检查所有用户信息
SELECT 
    usename,
    usesuper,
    usecreatedb,
    usecreaterole
FROM pg_user
ORDER BY usename;

-- 8. 检查当前用户的角色成员关系
SELECT 
    r.rolname as role_name,
    m.rolname as member_name
FROM pg_roles r 
JOIN pg_auth_members am ON r.oid = am.roleid
JOIN pg_roles m ON am.member = m.oid
WHERE m.rolname = current_user;

-- 9. 显示所有schema
SELECT schema_name FROM information_schema.schemata ORDER BY schema_name; 