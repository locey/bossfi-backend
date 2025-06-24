@echo off
echo ===================================
echo BossFi 数据库创建脚本
echo ===================================

set PGPATH="E:\software\postgreSql\bin"
set PGUSER=postgres
set PGHOST=localhost
set PGPORT=5432

echo 1. 创建数据库和用户...
%PGPATH%\psql.exe -U %PGUSER% -h %PGHOST% -p %PGPORT% -c "DROP DATABASE IF EXISTS bossfi;"
%PGPATH%\psql.exe -U %PGUSER% -h %PGHOST% -p %PGPORT% -c "DROP USER IF EXISTS bossfier;"
%PGPATH%\psql.exe -U %PGUSER% -h %PGHOST% -p %PGPORT% -c "CREATE USER bossfier WITH PASSWORD 'bossfier' CREATEDB LOGIN;"
%PGPATH%\psql.exe -U %PGUSER% -h %PGHOST% -p %PGPORT% -c "CREATE DATABASE bossfi OWNER bossfier;"

echo 2. 执行权限刷新脚本...
%PGPATH%\psql.exe -U %PGUSER% -h %PGHOST% -p %PGPORT% -f "scripts/refresh_permissions.sql"

echo 3. 创建数据库表结构...
%PGPATH%\psql.exe -U %PGUSER% -h %PGHOST% -p %PGPORT% -d bossfi -f "migrations/001_init_postgres.sql"

echo 4. 验证表创建结果...
%PGPATH%\psql.exe -U %PGUSER% -h %PGHOST% -p %PGPORT% -d bossfi -c "\dt"

echo ===================================
echo 数据库创建完成！
echo ===================================
echo 数据库信息:
echo 数据库名: bossfi
echo 用户名: bossfier
echo 密码: bossfier
echo 主机: localhost
echo 端口: 5432
echo ===================================

pause 