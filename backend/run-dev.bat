@echo off
cd /d "%~dp0"
rem 本机 Go 为 32 位安装时需指定 amd64，否则 excelize 等依赖无法编译
set GOARCH=amd64
rem 奇安信 AI 网关（Anthropic Messages 格式）
set DEEPSEEK_API_BASE=https://ai-gateway.qianxin-inc.cn
set DEEPSEEK_MODEL=glm-5.1
set DEEPSEEK_API_KEY=sk-uPYVy3tvrO5k8o5EQu4gyw
go run ./cmd/server
