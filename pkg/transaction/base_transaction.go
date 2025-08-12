package transaction

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BASETransactionImpl BASE事务实现（最终一致性）
type BASETransactionImpl struct {
	id            string
	status        TransactionStatus
	operations    []BASEOperation
	compensations []BASECompensation
	createdAt     time.Time
	updatedAt     time.Time
	timeout       time.Duration
	mu            sync.RWMutex
}

// BASEOperation BASE操作
type BASEOperation struct {
	ID          string
	Type        string
	SQL         string
	DataSource  string
	Parameters  []interface{}
	Status      string
	RetryCount  int
	MaxRetries  int
	ExecutedAt  *time.Time
	Error       error
}

// BASECompensation BASE补偿操作
type BASECompensation struct {
	ID          string
	OperationID string
	SQL         string
	DataSource  string
	Parameters  []interface{}
	Status      string
	ExecutedAt  *time.Time
	Error       error
}

// NewBASETransaction 创建BASE事务
func NewBASETransaction(id string) *BASETransactionImpl {
	return &BASETransactionImpl{
		id:            id,
		status:        StatusActive,
		operations:    make([]BASEOperation, 0),
		compensations: make([]BASECompensation, 0),
		createdAt:     time.Now(),
		updatedAt:     time.Now(),
		timeout:       30 * time.Minute, // 默认30分钟超时
	}
}

// Begin 开始BASE事务
func (t *BASETransactionImpl) Begin(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.status != StatusActive {
		return fmt.Errorf("BASE transaction %s is not in active status", t.id)
	}

	return nil
}

// Commit 提交BASE事务
func (t *BASETransactionImpl) Commit(ctx context.Context) error {
	t.mu.Lock()
	if t.status != StatusActive {
		t.mu.Unlock()
		return fmt.Errorf("BASE transaction %s is not in active status", t.id)
	}

	t.status = StatusPrepared // 使用StatusPrepared表示正在提交
	t.updatedAt = time.Now()
	t.mu.Unlock()

	// 异步执行操作
	go t.executeOperations(ctx)

	return nil
}

// Rollback 回滚BASE事务
func (t *BASETransactionImpl) Rollback(ctx context.Context) error {
	t.mu.Lock()
	if t.status == StatusCommitted {
		t.mu.Unlock()
		return fmt.Errorf("BASE transaction %s is already committed", t.id)
	}

	t.status = StatusRolledBack
	t.updatedAt = time.Now()
	t.mu.Unlock()

	// 异步执行补偿
	go t.executeCompensations(ctx)

	return nil
}

// GetStatus 获取事务状态
func (t *BASETransactionImpl) GetStatus() TransactionStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

// GetID 获取事务ID
func (t *BASETransactionImpl) GetID() string {
	return t.id
}

// GetType 获取事务类型
func (t *BASETransactionImpl) GetType() TransactionType {
	return BaseTransaction
}

// AddOperation 添加操作
func (t *BASETransactionImpl) AddOperation(op BASEOperation) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.status != StatusActive {
		return fmt.Errorf("cannot add operation to transaction in status %v", t.status)
	}

	op.ID = fmt.Sprintf("op_%d", time.Now().UnixNano())
	op.Status = "PENDING"
	op.MaxRetries = 3

	t.operations = append(t.operations, op)
	t.updatedAt = time.Now()

	return nil
}

// AddCompensation 添加补偿操作
func (t *BASETransactionImpl) AddCompensation(comp BASECompensation) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.status != StatusActive {
		return fmt.Errorf("cannot add compensation to transaction in status %v", t.status)
	}

	comp.ID = fmt.Sprintf("comp_%d", time.Now().UnixNano())
	comp.Status = "PENDING"

	t.compensations = append(t.compensations, comp)
	t.updatedAt = time.Now()

	return nil
}

// executeOperations 执行操作
func (t *BASETransactionImpl) executeOperations(ctx context.Context) {
	success := true

	// 执行所有操作
	for i := range t.operations {
		op := &t.operations[i]
		if err := t.executeOperation(ctx, op); err != nil {
			success = false
			break
		}
	}

	t.mu.Lock()
	if success {
		t.status = StatusCommitted
	} else {
		t.status = StatusFailed
	}
	t.updatedAt = time.Now()
	t.mu.Unlock()

	// 如果失败，执行补偿
	if !success {
		t.executeCompensations(ctx)
	}
}

// executeOperation 执行单个操作
func (t *BASETransactionImpl) executeOperation(ctx context.Context, op *BASEOperation) error {
	for op.RetryCount <= op.MaxRetries {
		op.Status = "EXECUTING"

		// 这里应该执行实际的SQL操作
		// 简化实现，实际需要根据DataSource执行SQL
		if op.DataSource != "" && op.SQL != "" {
			// 模拟执行成功
			op.Status = "COMPLETED"
			now := time.Now()
			op.ExecutedAt = &now
			return nil
		}

		op.RetryCount++
		if op.RetryCount <= op.MaxRetries {
			op.Status = "RETRYING"
			time.Sleep(time.Duration(op.RetryCount) * time.Second)
		}
	}

	op.Status = "FAILED"
	return fmt.Errorf("operation failed after %d retries", op.MaxRetries)
}

// executeCompensations 执行补偿操作
func (t *BASETransactionImpl) executeCompensations(ctx context.Context) {
	// 逆序执行补偿操作
	for i := len(t.compensations) - 1; i >= 0; i-- {
		comp := &t.compensations[i]
		comp.Status = "EXECUTING"

		// 这里应该执行实际的补偿SQL操作
		// 简化实现
		comp.Status = "COMPLETED"
		now := time.Now()
		comp.ExecutedAt = &now
	}

	t.mu.Lock()
	t.status = StatusRolledBack
	t.updatedAt = time.Now()
	t.mu.Unlock()
}

// GetOperations 获取所有操作
func (t *BASETransactionImpl) GetOperations() []BASEOperation {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	operations := make([]BASEOperation, len(t.operations))
	copy(operations, t.operations)
	return operations
}

// GetCompensations 获取所有补偿操作
func (t *BASETransactionImpl) GetCompensations() []BASECompensation {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	compensations := make([]BASECompensation, len(t.compensations))
	copy(compensations, t.compensations)
	return compensations
}

// IsExpired 检查事务是否过期
func (t *BASETransactionImpl) IsExpired() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	return time.Since(t.createdAt) > t.timeout
}

// SetTimeout 设置超时时间
func (t *BASETransactionImpl) SetTimeout(timeout time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.timeout = timeout
}