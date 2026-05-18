-- Worklog 本地开发库与用户（与 docker-compose / README 一致）
CREATE DATABASE IF NOT EXISTS worklog
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

CREATE USER IF NOT EXISTS 'worklog'@'localhost' IDENTIFIED BY 'worklog';
CREATE USER IF NOT EXISTS 'worklog'@'127.0.0.1' IDENTIFIED BY 'worklog';

GRANT ALL PRIVILEGES ON worklog.* TO 'worklog'@'localhost';
GRANT ALL PRIVILEGES ON worklog.* TO 'worklog'@'127.0.0.1';

FLUSH PRIVILEGES;

SELECT 'worklog database and user ready' AS status;
