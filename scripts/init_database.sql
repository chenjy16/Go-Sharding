-- 创建数据库
CREATE DATABASE IF NOT EXISTS ds_0 CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS ds_1 CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 使用 ds_0 数据库
USE ds_0;

-- 创建用户表分片
CREATE TABLE IF NOT EXISTS t_user (
    user_id BIGINT PRIMARY KEY,
    user_name VARCHAR(100) NOT NULL,
    user_email VARCHAR(200) NOT NULL,
    user_age INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_name (user_name),
    INDEX idx_user_email (user_email)
);

-- 创建订单表分片
CREATE TABLE IF NOT EXISTS t_order_0 (
    order_id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    order_status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    order_amount DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_order_status (order_status),
    INDEX idx_created_at (created_at)
);

CREATE TABLE IF NOT EXISTS t_order_1 (
    order_id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    order_status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    order_amount DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_order_status (order_status),
    INDEX idx_created_at (created_at)
);

-- 创建订单项表分片
CREATE TABLE IF NOT EXISTS t_order_item_0 (
    item_id BIGINT PRIMARY KEY,
    order_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    item_name VARCHAR(200) NOT NULL,
    item_price DECIMAL(10,2) NOT NULL,
    item_quantity INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_order_id (order_id),
    INDEX idx_user_id (user_id)
);

CREATE TABLE IF NOT EXISTS t_order_item_1 (
    item_id BIGINT PRIMARY KEY,
    order_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    item_name VARCHAR(200) NOT NULL,
    item_price DECIMAL(10,2) NOT NULL,
    item_quantity INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_order_id (order_id),
    INDEX idx_user_id (user_id)
);

-- 使用 ds_1 数据库
USE ds_1;

-- 创建用户表分片
CREATE TABLE IF NOT EXISTS t_user (
    user_id BIGINT PRIMARY KEY,
    user_name VARCHAR(100) NOT NULL,
    user_email VARCHAR(200) NOT NULL,
    user_age INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_name (user_name),
    INDEX idx_user_email (user_email)
);

-- 创建订单表分片
CREATE TABLE IF NOT EXISTS t_order_0 (
    order_id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    order_status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    order_amount DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_order_status (order_status),
    INDEX idx_created_at (created_at)
);

CREATE TABLE IF NOT EXISTS t_order_1 (
    order_id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    order_status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    order_amount DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_order_status (order_status),
    INDEX idx_created_at (created_at)
);

-- 创建订单项表分片
CREATE TABLE IF NOT EXISTS t_order_item_0 (
    item_id BIGINT PRIMARY KEY,
    order_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    item_name VARCHAR(200) NOT NULL,
    item_price DECIMAL(10,2) NOT NULL,
    item_quantity INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_order_id (order_id),
    INDEX idx_user_id (user_id)
);

CREATE TABLE IF NOT EXISTS t_order_item_1 (
    item_id BIGINT PRIMARY KEY,
    order_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    item_name VARCHAR(200) NOT NULL,
    item_price DECIMAL(10,2) NOT NULL,
    item_quantity INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_order_id (order_id),
    INDEX idx_user_id (user_id)
);

-- 插入测试数据到 ds_0
USE ds_0;

-- 用户数据（user_id 为偶数的用户）
INSERT IGNORE INTO t_user (user_id, user_name, user_email, user_age) VALUES
(2, '张三', 'zhangsan@example.com', 25),
(4, '王五', 'wangwu@example.com', 30),
(6, '赵六', 'zhaoliu@example.com', 28),
(8, '孙七', 'sunqi@example.com', 35);

-- 订单数据
INSERT IGNORE INTO t_order_0 (order_id, user_id, order_status, order_amount) VALUES
(1000, 2, 'PENDING', 299.99),
(1002, 4, 'COMPLETED', 199.50),
(1004, 6, 'PENDING', 399.00);

INSERT IGNORE INTO t_order_1 (order_id, user_id, order_status, order_amount) VALUES
(1001, 2, 'COMPLETED', 150.75),
(1003, 4, 'PENDING', 89.99),
(1005, 8, 'COMPLETED', 599.00);

-- 订单项数据
INSERT IGNORE INTO t_order_item_0 (item_id, order_id, user_id, item_name, item_price, item_quantity) VALUES
(10000, 1000, 2, '商品A', 99.99, 3),
(10002, 1002, 4, '商品C', 199.50, 1),
(10004, 1004, 6, '商品E', 399.00, 1);

INSERT IGNORE INTO t_order_item_1 (item_id, order_id, user_id, item_name, item_price, item_quantity) VALUES
(10001, 1001, 2, '商品B', 150.75, 1),
(10003, 1003, 4, '商品D', 89.99, 1),
(10005, 1005, 8, '商品F', 599.00, 1);

-- 插入测试数据到 ds_1
USE ds_1;

-- 用户数据（user_id 为奇数的用户）
INSERT IGNORE INTO t_user (user_id, user_name, user_email, user_age) VALUES
(1, '李四', 'lisi@example.com', 27),
(3, '陈五', 'chenwu@example.com', 32),
(5, '刘六', 'liuliu@example.com', 29),
(7, '周七', 'zhouqi@example.com', 26);

-- 订单数据
INSERT IGNORE INTO t_order_0 (order_id, user_id, order_status, order_amount) VALUES
(2000, 1, 'PENDING', 199.99),
(2002, 3, 'COMPLETED', 299.50),
(2004, 5, 'PENDING', 149.00);

INSERT IGNORE INTO t_order_1 (order_id, user_id, order_status, order_amount) VALUES
(2001, 1, 'COMPLETED', 89.75),
(2003, 3, 'PENDING', 199.99),
(2005, 7, 'COMPLETED', 399.00);

-- 订单项数据
INSERT IGNORE INTO t_order_item_0 (item_id, order_id, user_id, item_name, item_price, item_quantity) VALUES
(20000, 2000, 1, '商品G', 99.99, 2),
(20002, 2002, 3, '商品I', 299.50, 1),
(20004, 2004, 5, '商品K', 149.00, 1);

INSERT IGNORE INTO t_order_item_1 (item_id, order_id, user_id, item_name, item_price, item_quantity) VALUES
(20001, 2001, 1, '商品H', 89.75, 1),
(20003, 2003, 3, '商品J', 199.99, 1),
(20005, 2005, 7, '商品L', 399.00, 1);