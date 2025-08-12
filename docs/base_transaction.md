# BASE事务实现

## 概述

BASE事务是一种最终一致性的分布式事务模式，它通过牺牲强一致性来获得更好的可用性和性能。BASE代表：
- **Basically Available（基本可用）**：系统在出现故障时仍能保证核心功能可用
- **Soft state（软状态）**：允许系统存在中间状态，不要求实时一致性
- **Eventually consistent（最终一致性）**：系统最终会达到一致状态

## 特性

我们的BASE事务实现提供以下特性：

### 1. 异步执行
- 事务提交后立即返回，操作在后台异步执行
- 支持操作重试机制，提高成功率
- 失败时自动执行补偿操作

### 2. 补偿机制
- 支持为每个操作定义补偿操作
- 失败时按逆序执行补偿操作
- 确保系统最终达到一致状态

### 3. 超时管理
- 支持事务超时设置
- 自动检测过期事务
- 防止长时间运行的事务占用资源

### 4. 状态管理
- 完整的事务状态跟踪
- 支持并发安全的状态更新
- 提供状态查询接口

## 使用方法

### 1. 创建事务管理器

```go
tm := transaction.NewTransactionManager()
defer tm.Close()
```

### 2. 开始BASE事务

```go
ctx := context.Background()
tx, err := tm.Begin(ctx, transaction.BaseTransaction)
if err != nil {
    log.Fatalf("Failed to begin BASE transaction: %v", err)
}

baseTx := tx.(*transaction.BASETransactionImpl)
```

### 3. 添加操作

```go
op := transaction.BASEOperation{
    Type:       "INSERT",
    SQL:        "INSERT INTO orders (user_id, amount) VALUES (?, ?)",
    DataSource: "order_db",
    Parameters: []interface{}{123, 99.99},
}

err := baseTx.AddOperation(op)
if err != nil {
    log.Fatalf("Failed to add operation: %v", err)
}
```

### 4. 添加补偿操作

```go
comp := transaction.BASECompensation{
    OperationID: "op1",
    SQL:         "DELETE FROM orders WHERE user_id = ? AND amount = ?",
    DataSource:  "order_db",
    Parameters:  []interface{}{123, 99.99},
}

err := baseTx.AddCompensation(comp)
if err != nil {
    log.Fatalf("Failed to add compensation: %v", err)
}
```

### 5. 提交事务

```go
err := baseTx.Commit(ctx)
if err != nil {
    log.Fatalf("Failed to commit transaction: %v", err)
}
```

### 6. 检查状态

```go
status := baseTx.GetStatus()
operations := baseTx.GetOperations()
compensations := baseTx.GetCompensations()
```

## 事务状态

BASE事务支持以下状态：

- **StatusActive (0)**：事务活跃状态，可以添加操作
- **StatusPrepared (1)**：事务正在执行中
- **StatusCommitted (2)**：事务成功提交
- **StatusRolledBack (3)**：事务已回滚
- **StatusFailed (4)**：事务执行失败

## 操作状态

每个操作都有自己的状态：

- **PENDING**：等待执行
- **EXECUTING**：正在执行
- **COMPLETED**：执行成功
- **RETRYING**：重试中
- **FAILED**：执行失败

## 最佳实践

### 1. 操作设计
- 保持操作的幂等性
- 设计合理的补偿操作
- 避免长时间运行的操作

### 2. 错误处理
- 合理设置重试次数
- 监控操作执行状态
- 及时处理失败的事务

### 3. 性能优化
- 合理设置超时时间
- 避免过多的操作在单个事务中
- 定期清理过期事务

## 示例

完整的使用示例请参考：`examples/base_transaction/main.go`

## 注意事项

1. BASE事务不保证强一致性，适用于对一致性要求不严格的场景
2. 补偿操作必须是幂等的，因为可能会被多次执行
3. 事务的最终一致性依赖于补偿机制的正确实现
4. 在高并发场景下，需要注意资源竞争和死锁问题

## 与其他事务类型的比较

| 特性 | LOCAL事务 | XA事务 | BASE事务 |
|------|-----------|--------|----------|
| 一致性 | 强一致性 | 强一致性 | 最终一致性 |
| 性能 | 高 | 中 | 高 |
| 可用性 | 中 | 低 | 高 |
| 复杂度 | 低 | 高 | 中 |
| 适用场景 | 单数据源 | 多数据源强一致性 | 多数据源最终一致性 |