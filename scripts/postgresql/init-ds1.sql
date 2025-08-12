-- PostgreSQL 数据源 1 初始化脚本

-- 启用必要的扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "btree_gin";
CREATE EXTENSION IF NOT EXISTS "btree_gist";

-- 创建分片表 user_0
CREATE TABLE IF NOT EXISTS user_0 (
    user_id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    phone VARCHAR(20),
    date_of_birth DATE,
    gender VARCHAR(10),
    address JSONB,
    preferences JSONB,
    tags TEXT[],
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE,
    search_vector TSVECTOR
);

-- 创建分片表 user_1
CREATE TABLE IF NOT EXISTS user_1 (
    user_id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    phone VARCHAR(20),
    date_of_birth DATE,
    gender VARCHAR(10),
    address JSONB,
    preferences JSONB,
    tags TEXT[],
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE,
    search_vector TSVECTOR
);

-- 创建分片表 order_0
CREATE TABLE IF NOT EXISTS order_0 (
    order_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    order_number VARCHAR(50) UNIQUE NOT NULL,
    product_name VARCHAR(200) NOT NULL,
    product_sku VARCHAR(100),
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price DECIMAL(10,2) NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    tax_amount DECIMAL(10,2) DEFAULT 0,
    shipping_amount DECIMAL(10,2) DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'CNY',
    status VARCHAR(20) DEFAULT 'pending',
    payment_method VARCHAR(50),
    payment_status VARCHAR(20) DEFAULT 'unpaid',
    shipping_address JSONB,
    billing_address JSONB,
    order_items JSONB,
    metadata JSONB,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    shipped_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE
);

-- 创建分片表 order_1
CREATE TABLE IF NOT EXISTS order_1 (
    order_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    order_number VARCHAR(50) UNIQUE NOT NULL,
    product_name VARCHAR(200) NOT NULL,
    product_sku VARCHAR(100),
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price DECIMAL(10,2) NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    tax_amount DECIMAL(10,2) DEFAULT 0,
    shipping_amount DECIMAL(10,2) DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'CNY',
    status VARCHAR(20) DEFAULT 'pending',
    payment_method VARCHAR(50),
    payment_status VARCHAR(20) DEFAULT 'unpaid',
    shipping_address JSONB,
    billing_address JSONB,
    order_items JSONB,
    metadata JSONB,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    shipped_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE
);

-- 创建分片表 order_2
CREATE TABLE IF NOT EXISTS order_2 (
    order_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    order_number VARCHAR(50) UNIQUE NOT NULL,
    product_name VARCHAR(200) NOT NULL,
    product_sku VARCHAR(100),
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price DECIMAL(10,2) NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    tax_amount DECIMAL(10,2) DEFAULT 0,
    shipping_amount DECIMAL(10,2) DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'CNY',
    status VARCHAR(20) DEFAULT 'pending',
    payment_method VARCHAR(50),
    payment_status VARCHAR(20) DEFAULT 'unpaid',
    shipping_address JSONB,
    billing_address JSONB,
    order_items JSONB,
    metadata JSONB,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    shipped_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE
);

-- 创建分片表 order_3
CREATE TABLE IF NOT EXISTS order_3 (
    order_id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    order_number VARCHAR(50) UNIQUE NOT NULL,
    product_name VARCHAR(200) NOT NULL,
    product_sku VARCHAR(100),
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price DECIMAL(10,2) NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    tax_amount DECIMAL(10,2) DEFAULT 0,
    shipping_amount DECIMAL(10,2) DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'CNY',
    status VARCHAR(20) DEFAULT 'pending',
    payment_method VARCHAR(50),
    payment_status VARCHAR(20) DEFAULT 'unpaid',
    shipping_address JSONB,
    billing_address JSONB,
    order_items JSONB,
    metadata JSONB,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    shipped_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE
);

-- 创建索引
-- user 表索引
CREATE INDEX IF NOT EXISTS idx_user_0_username ON user_0(username);
CREATE INDEX IF NOT EXISTS idx_user_0_email ON user_0(email);
CREATE INDEX IF NOT EXISTS idx_user_0_status ON user_0(status);
CREATE INDEX IF NOT EXISTS idx_user_0_created_at ON user_0(created_at);
CREATE INDEX IF NOT EXISTS idx_user_0_search_vector ON user_0 USING gin(search_vector);
CREATE INDEX IF NOT EXISTS idx_user_0_tags ON user_0 USING gin(tags);
CREATE INDEX IF NOT EXISTS idx_user_0_address ON user_0 USING gin(address);

CREATE INDEX IF NOT EXISTS idx_user_1_username ON user_1(username);
CREATE INDEX IF NOT EXISTS idx_user_1_email ON user_1(email);
CREATE INDEX IF NOT EXISTS idx_user_1_status ON user_1(status);
CREATE INDEX IF NOT EXISTS idx_user_1_created_at ON user_1(created_at);
CREATE INDEX IF NOT EXISTS idx_user_1_search_vector ON user_1 USING gin(search_vector);
CREATE INDEX IF NOT EXISTS idx_user_1_tags ON user_1 USING gin(tags);
CREATE INDEX IF NOT EXISTS idx_user_1_address ON user_1 USING gin(address);

-- order 表索引
CREATE INDEX IF NOT EXISTS idx_order_0_user_id ON order_0(user_id);
CREATE INDEX IF NOT EXISTS idx_order_0_order_number ON order_0(order_number);
CREATE INDEX IF NOT EXISTS idx_order_0_status ON order_0(status);
CREATE INDEX IF NOT EXISTS idx_order_0_created_at ON order_0(created_at);
CREATE INDEX IF NOT EXISTS idx_order_0_total_amount ON order_0(total_amount);
CREATE INDEX IF NOT EXISTS idx_order_0_payment_status ON order_0(payment_status);

CREATE INDEX IF NOT EXISTS idx_order_1_user_id ON order_1(user_id);
CREATE INDEX IF NOT EXISTS idx_order_1_order_number ON order_1(order_number);
CREATE INDEX IF NOT EXISTS idx_order_1_status ON order_1(status);
CREATE INDEX IF NOT EXISTS idx_order_1_created_at ON order_1(created_at);
CREATE INDEX IF NOT EXISTS idx_order_1_total_amount ON order_1(total_amount);
CREATE INDEX IF NOT EXISTS idx_order_1_payment_status ON order_1(payment_status);

CREATE INDEX IF NOT EXISTS idx_order_2_user_id ON order_2(user_id);
CREATE INDEX IF NOT EXISTS idx_order_2_order_number ON order_2(order_number);
CREATE INDEX IF NOT EXISTS idx_order_2_status ON order_2(status);
CREATE INDEX IF NOT EXISTS idx_order_2_created_at ON order_2(created_at);
CREATE INDEX IF NOT EXISTS idx_order_2_total_amount ON order_2(total_amount);
CREATE INDEX IF NOT EXISTS idx_order_2_payment_status ON order_2(payment_status);

CREATE INDEX IF NOT EXISTS idx_order_3_user_id ON order_3(user_id);
CREATE INDEX IF NOT EXISTS idx_order_3_order_number ON order_3(order_number);
CREATE INDEX IF NOT EXISTS idx_order_3_status ON order_3(status);
CREATE INDEX IF NOT EXISTS idx_order_3_created_at ON order_3(created_at);
CREATE INDEX IF NOT EXISTS idx_order_3_total_amount ON order_3(total_amount);
CREATE INDEX IF NOT EXISTS idx_order_3_payment_status ON order_3(payment_status);

-- 创建触发器函数用于更新 updated_at 字段
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 创建触发器函数用于更新搜索向量
CREATE OR REPLACE FUNCTION update_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector = to_tsvector('english', 
        COALESCE(NEW.username, '') || ' ' ||
        COALESCE(NEW.email, '') || ' ' ||
        COALESCE(NEW.first_name, '') || ' ' ||
        COALESCE(NEW.last_name, '')
    );
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为 user 表创建触发器
CREATE TRIGGER update_user_0_updated_at BEFORE UPDATE ON user_0 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_0_search_vector BEFORE INSERT OR UPDATE ON user_0 
    FOR EACH ROW EXECUTE FUNCTION update_search_vector();

CREATE TRIGGER update_user_1_updated_at BEFORE UPDATE ON user_1 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_1_search_vector BEFORE INSERT OR UPDATE ON user_1 
    FOR EACH ROW EXECUTE FUNCTION update_search_vector();

-- 为 order 表创建触发器
CREATE TRIGGER update_order_0_updated_at BEFORE UPDATE ON order_0 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_order_1_updated_at BEFORE UPDATE ON order_1 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_order_2_updated_at BEFORE UPDATE ON order_2 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_order_3_updated_at BEFORE UPDATE ON order_3 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 插入测试数据（数据源 1 的数据）
-- 用户数据
INSERT INTO user_0 (username, email, password_hash, first_name, last_name, phone, address, tags) VALUES
('eve_0', 'eve0@example.com', '$2a$10$hash5', 'Eve', 'Davis', '+1234567894', 
 '{"street": "555 Maple St", "city": "Hangzhou", "country": "China"}', 
 ARRAY['premium', 'web']),
('grace_0', 'grace0@example.com', '$2a$10$hash7', 'Grace', 'Miller', '+1234567896', 
 '{"street": "777 Cedar St", "city": "Nanjing", "country": "China"}', 
 ARRAY['regular', 'mobile']);

INSERT INTO user_1 (username, email, password_hash, first_name, last_name, phone, address, tags) VALUES
('frank_1', 'frank1@example.com', '$2a$10$hash6', 'Frank', 'Wilson', '+1234567895', 
 '{"street": "666 Birch St", "city": "Wuhan", "country": "China"}', 
 ARRAY['vip', 'premium']),
('henry_1', 'henry1@example.com', '$2a$10$hash8', 'Henry', 'Taylor', '+1234567897', 
 '{"street": "888 Spruce St", "city": "Chengdu", "country": "China"}', 
 ARRAY['regular', 'web']);

-- 订单数据
INSERT INTO order_0 (user_id, order_number, product_name, product_sku, quantity, unit_price, total_amount, status, payment_status, order_items) VALUES
(5, 'ORD-2024-000009', 'MacBook Air M2', 'MBA-M2-001', 1, 8999.00, 8999.00, 'completed', 'paid', 
 '[{"name": "MacBook Air M2", "sku": "MBA-M2-001", "quantity": 1, "price": 8999.00}]'),
(7, 'ORD-2024-000011', 'Apple Pencil', 'AP-001', 1, 899.00, 899.00, 'shipped', 'paid', 
 '[{"name": "Apple Pencil", "sku": "AP-001", "quantity": 1, "price": 899.00}]');

INSERT INTO order_1 (user_id, order_number, product_name, product_sku, quantity, unit_price, total_amount, status, payment_status, order_items) VALUES
(6, 'ORD-2024-000010', 'iPhone 15', 'IP15-001', 1, 5999.00, 5999.00, 'pending', 'unpaid', 
 '[{"name": "iPhone 15", "sku": "IP15-001", "quantity": 1, "price": 5999.00}]'),
(8, 'ORD-2024-000012', 'iPad Pro 12.9"', 'IPP-129-001', 1, 8599.00, 8599.00, 'processing', 'paid', 
 '[{"name": "iPad Pro 12.9\"", "sku": "IPP-129-001", "quantity": 1, "price": 8599.00}]');

INSERT INTO order_2 (user_id, order_number, product_name, product_sku, quantity, unit_price, total_amount, status, payment_status, order_items) VALUES
(5, 'ORD-2024-000013', 'AirTag 4-pack', 'AT-4P-001', 1, 799.00, 799.00, 'completed', 'paid', 
 '[{"name": "AirTag 4-pack", "sku": "AT-4P-001", "quantity": 1, "price": 799.00}]'),
(7, 'ORD-2024-000015', 'MagSafe Charger', 'MSC-001', 2, 329.00, 658.00, 'shipped', 'paid', 
 '[{"name": "MagSafe Charger", "sku": "MSC-001", "quantity": 2, "price": 329.00}]');

INSERT INTO order_3 (user_id, order_number, product_name, product_sku, quantity, unit_price, total_amount, status, payment_status, order_items) VALUES
(6, 'ORD-2024-000014', 'HomePod mini', 'HPM-001', 2, 749.00, 1498.00, 'cancelled', 'refunded', 
 '[{"name": "HomePod mini", "sku": "HPM-001", "quantity": 2, "price": 749.00}]'),
(8, 'ORD-2024-000016', 'Apple TV 4K', 'ATV4K-001', 1, 1499.00, 1499.00, 'delivered', 'paid', 
 '[{"name": "Apple TV 4K", "sku": "ATV4K-001", "quantity": 1, "price": 1499.00}]');

-- 创建视图
CREATE OR REPLACE VIEW user_order_summary AS
SELECT 
    u.user_id,
    u.username,
    u.email,
    u.status as user_status,
    COUNT(o.order_id) as total_orders,
    COALESCE(SUM(o.total_amount), 0) as total_spent,
    COALESCE(AVG(o.total_amount), 0) as avg_order_value,
    MAX(o.created_at) as last_order_date
FROM (
    SELECT * FROM user_0
    UNION ALL
    SELECT * FROM user_1
) u
LEFT JOIN (
    SELECT * FROM order_0
    UNION ALL
    SELECT * FROM order_1
    UNION ALL
    SELECT * FROM order_2
    UNION ALL
    SELECT * FROM order_3
) o ON u.user_id = o.user_id
GROUP BY u.user_id, u.username, u.email, u.status;

-- 创建函数
CREATE OR REPLACE FUNCTION get_user_orders(p_user_id BIGINT)
RETURNS TABLE(
    order_id BIGINT,
    order_number VARCHAR(50),
    product_name VARCHAR(200),
    total_amount DECIMAL(10,2),
    status VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT o.order_id, o.order_number, o.product_name, o.total_amount, o.status, o.created_at
    FROM (
        SELECT * FROM order_0 WHERE user_id = p_user_id
        UNION ALL
        SELECT * FROM order_1 WHERE user_id = p_user_id
        UNION ALL
        SELECT * FROM order_2 WHERE user_id = p_user_id
        UNION ALL
        SELECT * FROM order_3 WHERE user_id = p_user_id
    ) o
    ORDER BY o.created_at DESC;
END;
$$ LANGUAGE plpgsql;

-- 创建存储过程
CREATE OR REPLACE FUNCTION create_order(
    p_user_id BIGINT,
    p_product_name VARCHAR(200),
    p_product_sku VARCHAR(100),
    p_quantity INTEGER,
    p_unit_price DECIMAL(10,2)
) RETURNS BIGINT AS $$
DECLARE
    v_order_id BIGINT;
    v_order_number VARCHAR(50);
    v_total_amount DECIMAL(10,2);
    v_table_suffix INTEGER;
BEGIN
    -- 生成订单号
    v_order_number := 'ORD-' || TO_CHAR(NOW(), 'YYYY-MM-DD-') || LPAD(NEXTVAL('order_number_seq')::TEXT, 6, '0');
    
    -- 计算总金额
    v_total_amount := p_quantity * p_unit_price;
    
    -- 根据订单ID确定分片表
    v_table_suffix := (EXTRACT(EPOCH FROM NOW())::BIGINT) % 4;
    
    -- 插入到对应的分片表
    CASE v_table_suffix
        WHEN 0 THEN
            INSERT INTO order_0 (user_id, order_number, product_name, product_sku, quantity, unit_price, total_amount)
            VALUES (p_user_id, v_order_number, p_product_name, p_product_sku, p_quantity, p_unit_price, v_total_amount)
            RETURNING order_id INTO v_order_id;
        WHEN 1 THEN
            INSERT INTO order_1 (user_id, order_number, product_name, product_sku, quantity, unit_price, total_amount)
            VALUES (p_user_id, v_order_number, p_product_name, p_product_sku, p_quantity, p_unit_price, v_total_amount)
            RETURNING order_id INTO v_order_id;
        WHEN 2 THEN
            INSERT INTO order_2 (user_id, order_number, product_name, product_sku, quantity, unit_price, total_amount)
            VALUES (p_user_id, v_order_number, p_product_name, p_product_sku, p_quantity, p_unit_price, v_total_amount)
            RETURNING order_id INTO v_order_id;
        WHEN 3 THEN
            INSERT INTO order_3 (user_id, order_number, product_name, product_sku, quantity, unit_price, total_amount)
            VALUES (p_user_id, v_order_number, p_product_name, p_product_sku, p_quantity, p_unit_price, v_total_amount)
            RETURNING order_id INTO v_order_id;
    END CASE;
    
    RETURN v_order_id;
END;
$$ LANGUAGE plpgsql;

-- 创建序列
CREATE SEQUENCE IF NOT EXISTS order_number_seq START 1000;

-- 授权
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO sharding_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO sharding_user;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO sharding_user;

-- 输出初始化完成信息
DO $$
BEGIN
    RAISE NOTICE 'PostgreSQL 数据源 1 初始化完成';
    RAISE NOTICE '已创建表: user_0, user_1, order_0, order_1, order_2, order_3';
    RAISE NOTICE '已创建索引、触发器、视图和函数';
    RAISE NOTICE '已插入测试数据';
END $$;