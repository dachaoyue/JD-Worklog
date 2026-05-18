@echo off
cd /d "%~dp0"
rem 本机 Go 为 32 位安装时需指定 amd64，否则 excelize 等依赖无法编译
set GOARCH=amd64
go run ./cmd/server
