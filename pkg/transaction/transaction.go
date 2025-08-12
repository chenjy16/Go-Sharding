package transaction

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// TransactionType 事务类型
type TransactionType int

const (
	// LocalTransaction 本地事务
	LocalTransaction TransactionType = iota
	// XATransaction XA 分布式事务
	XATransaction
	// BaseTransaction BASE 柔性事务
	BaseTransaction
)

// TransactionStatus 事务状态
type TransactionStatus int

const (
	// StatusActive 活跃状态
	StatusActive TransactionStatus = iota
	// StatusPrepared 已准备状态（XA 事务）
	StatusPrepared
	// StatusCommitted 已提交状态
	StatusCommitted
	// StatusRolledBack 已回滚状态
	StatusRolledBack
	// StatusFailed 失败状态
	StatusFailed
)

// Transaction 事务接口
type Transaction interface {
	// Begin 开始事务
	Begin(ctx context.Context) error
	// Commit 提交事务
	Commit(ctx context.Context) error
	// Rollback 回滚事务
	Rollback(ctx context.Context) error
	// GetStatus 获取事务状态
	GetStatus() TransactionStatus
	// GetID 获取事务 ID
	GetID() string
	// GetType 获取事务类型
	GetType() TransactionType
}

// TransactionManager 事务管理器接口
type TransactionManager interface {
	// Begin 开始事务
	Begin(ctx context.Context, txType TransactionType) (Transaction, error)
	// GetTransaction 获取当前事务
	GetTransaction(ctx context.Context) Transaction
	// RegisterDataSource 注册数据源
	RegisterDataSource(name string, db *sql.DB) error
	// Close 关闭事务管理器
	Close() error
}

// LocalTransactionImpl 本地事务实现
type LocalTransactionImpl struct {
	id       string
	status   TransactionStatus
	tx       *sql.Tx
	db       *sql.DB
	mu       sync.RWMutex
	startTime time.Time
}

// NewLocalTransaction 创建本地事务
func NewLocalTransaction(id string, db *sql.DB) *LocalTransactionImpl {
	return &LocalTransactionImpl{
		id:        id,
		status:    StatusActive,
		db:        db,
		startTime: time.Now(),
	}
}

// Begin 开始本地事务
func (t *LocalTransactionImpl) Begin(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.status != StatusActive {
		return fmt.Errorf("transaction %s is not in active status", t.id)
	}

	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		t.status = StatusFailed
		return fmt.Errorf("failed to begin transaction %s: %w", t.id, err)
	}

	t.tx = tx
	return nil
}

// Commit 提交本地事务
func (t *LocalTransactionImpl) Commit(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.tx == nil {
		return fmt.Errorf("transaction %s is not started", t.id)
	}

	if t.status != StatusActive {
		return fmt.Errorf("transaction %s is not in active status", t.id)
	}

	err := t.tx.Commit()
	if err != nil {
		t.status = StatusFailed
		return fmt.Errorf("failed to commit transaction %s: %w", t.id, err)
	}

	t.status = StatusCommitted
	return nil
}

// Rollback 回滚本地事务
func (t *LocalTransactionImpl) Rollback(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.tx == nil {
		return fmt.Errorf("transaction %s is not started", t.id)
	}

	if t.status == StatusCommitted {
		return fmt.Errorf("transaction %s is already committed", t.id)
	}

	err := t.tx.Rollback()
	if err != nil {
		t.status = StatusFailed
		return fmt.Errorf("failed to rollback transaction %s: %w", t.id, err)
	}

	t.status = StatusRolledBack
	return nil
}

// GetStatus 获取事务状态
func (t *LocalTransactionImpl) GetStatus() TransactionStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

// GetID 获取事务 ID
func (t *LocalTransactionImpl) GetID() string {
	return t.id
}

// GetType 获取事务类型
func (t *LocalTransactionImpl) GetType() TransactionType {
	return LocalTransaction
}

// GetTx 获取底层事务对象
func (t *LocalTransactionImpl) GetTx() *sql.Tx {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.tx
}

// XATransactionImpl XA 分布式事务实现
type XATransactionImpl struct {
	id          string
	status      TransactionStatus
	branches    map[string]*XABranch
	mu          sync.RWMutex
	startTime   time.Time
	coordinator *XACoordinator
}

// XABranch XA 事务分支
type XABranch struct {
	ID         string
	DataSource string
	Tx         *sql.Tx
	Status     TransactionStatus
}

// XACoordinator XA 事务协调器
type XACoordinator struct {
	transactions map[string]*XATransactionImpl
	mu           sync.RWMutex
}

// NewXATransaction 创建 XA 事务
func NewXATransaction(id string, coordinator *XACoordinator) *XATransactionImpl {
	return &XATransactionImpl{
		id:          id,
		status:      StatusActive,
		branches:    make(map[string]*XABranch),
		startTime:   time.Now(),
		coordinator: coordinator,
	}
}

// Begin 开始 XA 事务
func (t *XATransactionImpl) Begin(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.status != StatusActive {
		return fmt.Errorf("XA transaction %s is not in active status", t.id)
	}

	// XA 事务的开始逻辑
	// 这里可以添加 XA START 的实现
	return nil
}

// Commit 提交 XA 事务（两阶段提交）
func (t *XATransactionImpl) Commit(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.status != StatusActive {
		return fmt.Errorf("XA transaction %s is not in active status", t.id)
	}

	// 第一阶段：准备
	for _, branch := range t.branches {
		if err := t.prepareBranch(ctx, branch); err != nil {
			// 如果准备失败，回滚所有分支
			t.rollbackAllBranches(ctx)
			t.status = StatusRolledBack
			return fmt.Errorf("failed to prepare branch %s: %w", branch.ID, err)
		}
	}

	t.status = StatusPrepared

	// 第二阶段：提交
	for _, branch := range t.branches {
		if err := t.commitBranch(ctx, branch); err != nil {
			// 提交失败，记录错误但继续尝试其他分支
			// 在实际实现中，这里需要更复杂的错误处理和恢复机制
			continue
		}
	}

	t.status = StatusCommitted
	return nil
}

// Rollback 回滚 XA 事务
func (t *XATransactionImpl) Rollback(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.status == StatusCommitted {
		return fmt.Errorf("XA transaction %s is already committed", t.id)
	}

	t.rollbackAllBranches(ctx)
	t.status = StatusRolledBack
	return nil
}

// prepareBranch 准备分支事务
func (t *XATransactionImpl) prepareBranch(ctx context.Context, branch *XABranch) error {
	// 这里应该执行 XA PREPARE 命令
	// 简化实现，实际需要调用数据库的 XA 接口
	branch.Status = StatusPrepared
	return nil
}

// commitBranch 提交分支事务
func (t *XATransactionImpl) commitBranch(ctx context.Context, branch *XABranch) error {
	// 这里应该执行 XA COMMIT 命令
	if branch.Tx != nil {
		err := branch.Tx.Commit()
		if err != nil {
			return err
		}
	}
	branch.Status = StatusCommitted
	return nil
}

// rollbackAllBranches 回滚所有分支事务
func (t *XATransactionImpl) rollbackAllBranches(ctx context.Context) {
	for _, branch := range t.branches {
		if branch.Tx != nil {
			branch.Tx.Rollback()
		}
		branch.Status = StatusRolledBack
	}
}

// GetStatus 获取事务状态
func (t *XATransactionImpl) GetStatus() TransactionStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

// GetID 获取事务 ID
func (t *XATransactionImpl) GetID() string {
	return t.id
}

// GetType 获取事务类型
func (t *XATransactionImpl) GetType() TransactionType {
	return XATransaction
}

// AddBranch 添加事务分支
func (t *XATransactionImpl) AddBranch(branchID, dataSource string, tx *sql.Tx) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.branches[branchID] = &XABranch{
		ID:         branchID,
		DataSource: dataSource,
		Tx:         tx,
		Status:     StatusActive,
	}
}

// TransactionManagerImpl 事务管理器实现
type TransactionManagerImpl struct {
	dataSources   map[string]*sql.DB
	transactions  map[string]Transaction
	xaCoordinator *XACoordinator
	mu            sync.RWMutex
}

// NewTransactionManager 创建事务管理器
func NewTransactionManager() *TransactionManagerImpl {
	return &TransactionManagerImpl{
		dataSources:  make(map[string]*sql.DB),
		transactions: make(map[string]Transaction),
		xaCoordinator: &XACoordinator{
			transactions: make(map[string]*XATransactionImpl),
		},
	}
}

// Begin 开始事务
func (tm *TransactionManagerImpl) Begin(ctx context.Context, txType TransactionType) (Transaction, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	txID := generateTransactionID()

	var tx Transaction
	var err error

	switch txType {
	case LocalTransaction:
		// 对于本地事务，使用第一个数据源
		var db *sql.DB
		for _, ds := range tm.dataSources {
			db = ds
			break
		}
		if db == nil {
			return nil, fmt.Errorf("no data source available for local transaction")
		}
		tx = NewLocalTransaction(txID, db)
	case XATransaction:
		tx = NewXATransaction(txID, tm.xaCoordinator)
	case BaseTransaction:
		// BASE 事务的实现
		return nil, fmt.Errorf("BASE transaction not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported transaction type: %v", txType)
	}

	err = tx.Begin(ctx)
	if err != nil {
		return nil, err
	}

	tm.transactions[txID] = tx
	return tx, nil
}

// GetTransaction 获取当前事务
func (tm *TransactionManagerImpl) GetTransaction(ctx context.Context) Transaction {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	// 从上下文中获取事务 ID
	if txID, ok := ctx.Value("transaction_id").(string); ok {
		return tm.transactions[txID]
	}
	return nil
}

// RegisterDataSource 注册数据源
func (tm *TransactionManagerImpl) RegisterDataSource(name string, db *sql.DB) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.dataSources[name] = db
	return nil
}

// Close 关闭事务管理器
func (tm *TransactionManagerImpl) Close() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// 关闭所有未完成的事务
	for _, tx := range tm.transactions {
		if tx.GetStatus() == StatusActive {
			tx.Rollback(context.Background())
		}
	}

	// 关闭所有数据源
	for _, db := range tm.dataSources {
		db.Close()
	}

	return nil
}

// generateTransactionID 生成事务 ID
func generateTransactionID() string {
	return fmt.Sprintf("tx_%d", time.Now().UnixNano())
}

// TransactionContext 事务上下文
type TransactionContext struct {
	TransactionID string
	Type          TransactionType
	StartTime     time.Time
	Timeout       time.Duration
}

// WithTransaction 在上下文中设置事务
func WithTransaction(ctx context.Context, tx Transaction) context.Context {
	return context.WithValue(ctx, "transaction_id", tx.GetID())
}

// GetTransactionFromContext 从上下文中获取事务 ID
func GetTransactionFromContext(ctx context.Context) string {
	if txID, ok := ctx.Value("transaction_id").(string); ok {
		return txID
	}
	return ""
}