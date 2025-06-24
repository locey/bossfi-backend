@echo off
chcp 65001 >nul
echo ===================================
echo PostgreSQL Permission Fix Script
echo ===================================

set PGPATH="E:\software\postgreSql\bin"
set PGUSER=postgres
set PGHOST=localhost
set PGPORT=5432

echo Diagnosing permissions issue...
echo.

echo 1. Checking current tables with postgres user...
%PGPATH%\psql.exe -U %PGUSER% -h %PGHOST% -p %PGPORT% -d bossfi -c "SELECT tablename, tableowner FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename;"

echo.
echo 2. Fixing permissions...
%PGPATH%\psql.exe -U %PGUSER% -h %PGHOST% -p %PGPORT% -d bossfi -f "scripts/fix_permissions.sql"

echo.
echo 3. Testing bossfier access...
%PGPATH%\psql.exe -U bossfier -h %PGHOST% -p %PGPORT% -d bossfi -c "\dt"

echo.
echo 4. Listing tables with bossfier user...
%PGPATH%\psql.exe -U bossfier -h %PGHOST% -p %PGPORT% -d bossfi -c "SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename;"

echo ===================================
echo Permission fix completed!
echo Now bossfier user should see all tables
echo ===================================

pause 