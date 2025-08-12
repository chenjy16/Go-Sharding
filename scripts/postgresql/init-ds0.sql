-- PostgreSQL 数据源 0 初始化脚本

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

-- 插入测试数据
-- 用户数据
INSERT INTO user_0 (username, email, password_hash, first_name, last_name, phone, address, tags) VALUES
('alice_0', 'alice0@example.com', '$2a$10$hash1', 'Alice', 'Smith', '+1234567890', 
 '{"street": "123 Main St", "city": "Beijing", "country": "China"}', 
 ARRAY['vip', 'premium']),
('charlie_0', 'charlie0@example.com', '$2a$10$hash3', 'Charlie', 'Brown', '+1234567892', 
 '{"street": "789 Oak St", "city": "Shanghai", "country": "China"}', 
 ARRAY['regular', 'mobile']);

INSERT INTO user_1 (username, email, password_hash, first_name, last_name, phone, address, tags) VALUES
('bob_1', 'bob1@example.com', '$2a$10$hash2', 'Bob', 'Johnson', '+1234567891', 
 '{"street": "456 Elm St", "city": "Guangzhou", "country": "China"}', 
 ARRAY['regular', 'web']),
('diana_1', 'diana1@example.com', '$2a$10$hash4', 'Diana', 'Wilson', '+1234567893', 
 '{"street": "321 Pine St", "city": "Shenzhen", "country": "China"}', 
 ARRAY['vip', 'mobile']);

-- 订单数据
INSERT INTO order_0 (user_id, order_number, product_name, product_sku, quantity, unit_price, total_amount, status, payment_status, order_items) VALUES
(1, 'ORD-2024-000001', 'MacBook Pro 16"', 'MBP16-001', 1, 16999.00, 16999.00, 'completed', 'paid', 
 '[{"name": "MacBook Pro 16\"", "sku": "MBP16-001", "quantity": 1, "price": 16999.00}]'),
(3, 'ORD-2024-000003', 'AirPods Pro', 'APP-001', 2, 1999.00, 3998.00, 'shipped', 'paid', 
 '[{"name": "AirPods Pro", "sku": "APP-001", "quantity": 2, "price": 1999.00}]');

INSERT INTO order_1 (user_id, order_number, product_name, product_sku, quantity, unit_price, total_amount, status, payment_status, order_items) VALUES
(2, 'ORD-2024-000002', 'iPhone 15 Pro', 'IP15P-001', 1, 8999.00, 8999.00, 'pending', 'unpaid', 
 '[{"name": "iPhone 15 Pro", "sku": "IP15P-001", "quantity": 1, "price": 8999.00}]'),
(4, 'ORD-2024-000004', 'iPad Air', 'IPA-001', 1, 4599.00, 4599.00, 'processing', 'paid', 
 '[{"name": "iPad Air", "sku": "IPA-001", "quantity": 1, "price": 4599.00}]');

INSERT INTO order_2 (user_id, order_number, product_name, product_sku, quantity, unit_price, total_amount, status, payment_status, order_items) VALUES
(1, 'ORD-2024-000005', 'Magic Mouse', 'MM-001', 1, 799.00, 799.00, 'completed', 'paid', 
 '[{"name": "Magic Mouse", "sku": "MM-001", "quantity": 1, "price": 799.00}]'),
(3, 'ORD-2024-000007', 'USB-C Cable', 'USBC-001', 3, 199.00, 597.00, 'shipped', 'paid', 
 '[{"name": "USB-C Cable", "sku": "USBC-001", "quantity": 3, "price": 199.00}]');

INSERT INTO order_3 (user_id, order_number, product_name, product_sku, quantity, unit_price, total_amount, status, payment_status, order_items) VALUES
(2, 'ORD-2024-000006', 'Magic Keyboard', 'MK-001', 1, 1299.00, 1299.00, 'cancelled', 'refunded', 
 '[{"name": "Magic Keyboard", "sku": "MK-001", "quantity": 1, "price": 1299.00}]'),
(4, 'ORD-2024-000008', 'Apple Watch', 'AW-001', 1, 2999.00, 2999.00, 'delivered', 'paid', 
 '[{"name": "Apple Watch", "sku": "AW-001", "quantity": 1, "price": 2999.00}]');

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
CREATE SEQUENCE IF NOT EXISTS order_number_seq START 1;

-- 授权
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO sharding_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO sharding_user;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO sharding_user;

-- 输出初始化完成信息
DO $$
BEGIN
    RAISE NOTICE 'PostgreSQL 数据源 0 初始化完成';
    RAISE NOTICE '已创建表: user_0, user_1, order_0, order_1, order_2, order_3';
    RAISE NOTICE '已创建索引、触发器、视图和函数';
    RAISE NOTICE '已插入测试数据';
END $$;