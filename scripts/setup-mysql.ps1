# 使用 MySQL root 创建 worklog 库与账号
# 账号: worklog  密码: worklog  数据库: worklog

$ErrorActionPreference = "Stop"
$RootDir = Split-Path -Parent $PSScriptRoot
$SqlFile = Join-Path $PSScriptRoot "setup-mysql.sql"

if (-not (Test-Path $SqlFile)) {
    Write-Error "找不到 $SqlFile"
}

$mysql = Get-Command mysql -ErrorAction SilentlyContinue
if (-not $mysql) {
    Write-Error "未找到 mysql 命令，请确认 MySQL 已安装并已加入 PATH"
}

$secure = Read-Host "请输入 MySQL root 密码" -AsSecureString
$bstr = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($secure)
$rootPass = [Runtime.InteropServices.Marshal]::PtrToStringAuto($bstr)
[Runtime.InteropServices.Marshal]::ZeroFreeBSTR($bstr)

$env:MYSQL_PWD = $rootPass
try {
    Get-Content -Raw $SqlFile | & mysql -u root -h 127.0.0.1
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

    Write-Host ""
    Write-Host "MySQL 已配置完成：" -ForegroundColor Green
    Write-Host "  数据库: worklog"
    Write-Host "  用户:   worklog"
    Write-Host "  密码:   worklog"
    Write-Host "  主机:   127.0.0.1:3306"
    Write-Host ""
    Write-Host "启动 backend 示例："
    Write-Host '  $env:DB_HOST="127.0.0.1"; $env:DB_PORT="3306"; $env:DB_USER="worklog"; $env:DB_PASS="worklog"; $env:DB_NAME="worklog"'
    Write-Host "  cd `"$RootDir\backend`""
    Write-Host "  go run ./cmd/server"
}
finally {
    Remove-Item Env:MYSQL_PWD -ErrorAction SilentlyContinue
}
