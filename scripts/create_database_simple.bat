@echo off
chcp 65001 >nul
echo ===================================
echo BossFi Database Creation Script
echo ===================================

set PGPATH="E:\software\postgreSql\bin"
set PGUSER=postgres
set PGHOST=localhost
set PGPORT=5432
set PGPASSWORD=postgres

echo 1. Creating database and user...
%PGPATH%\psql.exe -U %PGUSER% -h %PGHOST% -p %PGPORT% -c "DROP DATABASE IF EXISTS bossfi;"
%PGPATH%\psql.exe -U %PGUSER% -h %PGHOST% -p %PGPORT% -c "DROP USER IF EXISTS bossfier;"
%PGPATH%\psql.exe -U %PGUSER% -h %PGHOST% -p %PGPORT% -c "CREATE USER bossfier WITH PASSWORD 'bossfier' CREATEDB LOGIN;"
%PGPATH%\psql.exe -U %PGUSER% -h %PGHOST% -p %PGPORT% -c "CREATE DATABASE bossfi OWNER bossfier;"
%PGPATH%\psql.exe -U %PGUSER% -h %PGHOST% -p %PGPORT% -c "GRANT ALL PRIVILEGES ON DATABASE bossfi TO bossfier;"

echo 2. Creating table structure...
%PGPATH%\psql.exe -U %PGUSER% -h %PGHOST% -p %PGPORT% -d bossfi -f "scripts/quick_setup.sql"

echo 3. Verifying tables...
%PGPATH%\psql.exe -U %PGUSER% -h %PGHOST% -p %PGPORT% -d bossfi -c "SELECT COUNT(*) as table_count FROM information_schema.tables WHERE table_schema = 'public';"

echo 4. Listing all tables...
%PGPATH%\psql.exe -U %PGUSER% -h %PGHOST% -p %PGPORT% -d bossfi -c "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name;"

echo ===================================
echo Database creation completed!
echo ===================================
echo Database info:
echo Database: bossfi
echo User: bossfier
echo Password: bossfier
echo Host: localhost
echo Port: 5432
echo ===================================

pause 