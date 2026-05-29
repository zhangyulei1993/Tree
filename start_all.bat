@echo off
set ROOT=D:\GitHub\zhangyulei\Tree

start "backend" cmd /k "cd /d %ROOT%\backend && go run ./cmd/server"

timeout /t 2 /nobreak >nul

start "admin" cmd /k "cd /d %ROOT%\admin && npm run dev"

start "web" cmd /k "cd /d %ROOT%\web && npm run dev"