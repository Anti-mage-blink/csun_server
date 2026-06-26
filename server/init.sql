-- ============================================
-- MySQL 初始化脚本
-- 仅在容器首次启动（data 目录为空）时自动执行
-- ============================================

-- 1. 创建通用数据库
CREATE DATABASE IF NOT EXISTS general
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_0900_ai_ci;

-- 2. 创建报价管理业务数据库
CREATE DATABASE IF NOT EXISTS quote_manage
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_0900_ai_ci;
