package transaction

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTransactionType_Constants(t *testing.T) {
	assert.Equal(t, TransactionType(0), LocalTransaction)
	assert.Equal(t, TransactionType(1), XATransaction)
	assert.Equal(t, TransactionType(2), BaseTransaction)
}

func TestTransactionStatus_Constants(t *testing.T) {
	assert.Equal(t, TransactionStatus(0), StatusActive)
	assert.Equal(t, TransactionStatus(1), StatusPrepared)
	assert.Equal(t, TransactionStatus(2), StatusCommitted)
	assert.Equal(t, TransactionStatus(3), StatusRolledBack)
	assert.Equal(t, TransactionStatus(4), StatusFailed)
}

func TestNewLocalTransaction(t *testing.T) {
	tx := NewLocalTransaction("test-tx-1", nil)
	assert.NotNil(t, tx)
	assert.Equal(t, "test-tx-1", tx.GetID())
	assert.Equal(t, LocalTransaction, tx.GetType())
	assert.Equal(t, StatusActive, tx.GetStatus())
}

func TestLocalTransaction_Begin(t *testing.T) {
	// 使用 nil DB 来测试基本功能，因为 Begin 方法会检查状态
	tx := NewLocalTransaction("test-tx-1", nil)
	
	// 首先将状态设置为非活跃状态来测试状态检查
	tx.status = StatusCommitted
	ctx := context.Background()

	// 测试非活跃状态下的 Begin
	err := tx.Begin(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in active status")
}

func TestLocalTransaction_Commit(t *testing.T) {
	tx := NewLocalTransaction("test-tx-1", nil)
	ctx := context.Background()

	// 测试未开始事务的提交
	err := tx.Commit(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not started")
}

func TestLocalTransaction_Rollback(t *testing.T) {
	tx := NewLocalTransaction("test-tx-1", nil)
	ctx := context.Background()

	// 测试未开始事务的回滚
	err := tx.Rollback(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not started")
}

func TestLocalTransaction_CommitWithoutBegin(t *testing.T) {
	tx := NewLocalTransaction("test-tx-1", nil)
	ctx := context.Background()

	err := tx.Commit(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not started")
}

func TestLocalTransaction_RollbackWithoutBegin(t *testing.T) {
	tx := NewLocalTransaction("test-tx-1", nil)
	ctx := context.Background()

	err := tx.Rollback(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not started")
}

func TestNewXATransaction(t *testing.T) {
	coordinator := &XACoordinator{
		transactions: make(map[string]*XATransactionImpl),
	}

	tx := NewXATransaction("xa-tx-1", coordinator)
	assert.NotNil(t, tx)
	assert.Equal(t, "xa-tx-1", tx.GetID())
	assert.Equal(t, XATransaction, tx.GetType())
	assert.Equal(t, StatusActive, tx.GetStatus())
	assert.NotNil(t, tx.branches)
}

func TestXATransaction_Begin(t *testing.T) {
	coordinator := &XACoordinator{
		transactions: make(map[string]*XATransactionImpl),
	}

	tx := NewXATransaction("xa-tx-1", coordinator)
	ctx := context.Background()

	err := tx.Begin(ctx)
	assert.NoError(t, err)
	assert.Equal(t, StatusActive, tx.GetStatus())
}

func TestXATransaction_AddBranch(t *testing.T) {
	coordinator := &XACoordinator{
		transactions: make(map[string]*XATransactionImpl),
	}

	tx := NewXATransaction("xa-tx-1", coordinator)

	// 模拟添加分支
	tx.AddBranch("branch-1", "ds1", nil)
	assert.Len(t, tx.branches, 1)
	assert.Contains(t, tx.branches, "branch-1")
	assert.Equal(t, "ds1", tx.branches["branch-1"].DataSource)
}

func TestXATransaction_Commit(t *testing.T) {
	coordinator := &XACoordinator{
		transactions: make(map[string]*XATransactionImpl),
	}

	tx := NewXATransaction("xa-tx-1", coordinator)
	ctx := context.Background()

	err := tx.Begin(ctx)
	assert.NoError(t, err)

	// 添加一些分支（不使用真实的数据库连接）
	tx.AddBranch("branch-1", "ds1", nil)
	tx.AddBranch("branch-2", "ds2", nil)

	err = tx.Commit(ctx)
	assert.NoError(t, err)
	assert.Equal(t, StatusCommitted, tx.GetStatus())
}

func TestXATransaction_Rollback(t *testing.T) {
	coordinator := &XACoordinator{
		transactions: make(map[string]*XATransactionImpl),
	}

	tx := NewXATransaction("xa-tx-1", coordinator)
	ctx := context.Background()

	err := tx.Begin(ctx)
	assert.NoError(t, err)

	tx.AddBranch("branch-1", "ds1", nil)

	err = tx.Rollback(ctx)
	assert.NoError(t, err)
	assert.Equal(t, StatusRolledBack, tx.GetStatus())
}

func TestNewTransactionManager(t *testing.T) {
	tm := NewTransactionManager()
	assert.NotNil(t, tm)
	assert.NotNil(t, tm.dataSources)
	assert.NotNil(t, tm.transactions)
	assert.NotNil(t, tm.xaCoordinator)
}

func TestTransactionManager_RegisterDataSource(t *testing.T) {
	tm := NewTransactionManager()
	
	// 使用 nil 数据库进行测试
	err := tm.RegisterDataSource("ds1", nil)
	assert.NoError(t, err)
	assert.Contains(t, tm.dataSources, "ds1")
}

func TestTransactionManager_BeginLocalTransaction(t *testing.T) {
	tm := NewTransactionManager()
	
	// 注册一个 nil 数据源
	err := tm.RegisterDataSource("ds1", nil)
	assert.NoError(t, err)

	ctx := context.Background()
	tx, err := tm.Begin(ctx, LocalTransaction)
	// 由于数据库为 nil，Begin 会失败
	assert.Error(t, err)
	assert.Nil(t, tx)
}

func TestTransactionManager_BeginXATransaction(t *testing.T) {
	tm := NewTransactionManager()
	ctx := context.Background()

	tx, err := tm.Begin(ctx, XATransaction)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, XATransaction, tx.GetType())
	assert.Equal(t, StatusActive, tx.GetStatus())
}

func TestTransactionManager_BeginBaseTransaction(t *testing.T) {
	tm := NewTransactionManager()
	ctx := context.Background()

	tx, err := tm.Begin(ctx, BaseTransaction)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "not implemented")
}

func TestTransactionManager_BeginLocalTransactionWithoutDataSource(t *testing.T) {
	tm := NewTransactionManager()
	ctx := context.Background()

	tx, err := tm.Begin(ctx, LocalTransaction)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "no data source available")
}

func TestTransactionManager_GetTransaction(t *testing.T) {
	tm := NewTransactionManager()
	
	ctx := context.Background()
	tx, err := tm.Begin(ctx, XATransaction)
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// 使用事务上下文
	ctxWithTx := WithTransaction(ctx, tx)
	retrievedTx := tm.GetTransaction(ctxWithTx)
	assert.NotNil(t, retrievedTx)
	assert.Equal(t, tx.GetID(), retrievedTx.GetID())
}

func TestTransactionManager_GetTransactionWithoutContext(t *testing.T) {
	tm := NewTransactionManager()
	ctx := context.Background()

	tx := tm.GetTransaction(ctx)
	assert.Nil(t, tx)
}

func TestTransactionManager_Close(t *testing.T) {
	tm := NewTransactionManager()

	ctx := context.Background()
	tx, err := tm.Begin(ctx, XATransaction)
	assert.NoError(t, err)
	assert.Equal(t, StatusActive, tx.GetStatus())

	err = tm.Close()
	assert.NoError(t, err)
}

func TestGenerateTransactionID(t *testing.T) {
	id1 := generateTransactionID()
	
	// 添加微小延迟确保时间戳不同
	time.Sleep(1 * time.Microsecond)
	
	id2 := generateTransactionID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "tx_")
	assert.Contains(t, id2, "tx_")
}

func TestWithTransaction(t *testing.T) {
	tx := NewLocalTransaction("test-tx-1", nil)
	ctx := context.Background()

	ctxWithTx := WithTransaction(ctx, tx)
	assert.NotNil(t, ctxWithTx)

	txID := GetTransactionFromContext(ctxWithTx)
	assert.Equal(t, "test-tx-1", txID)
}

func TestGetTransactionFromContext(t *testing.T) {
	ctx := context.Background()
	txID := GetTransactionFromContext(ctx)
	assert.Empty(t, txID)

	ctxWithTx := context.WithValue(ctx, "transaction_id", "test-tx-1")
	txID = GetTransactionFromContext(ctxWithTx)
	assert.Equal(t, "test-tx-1", txID)
}

func TestLocalTransaction_Concurrent(t *testing.T) {
	var wg sync.WaitGroup
	goroutines := 10

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			tx := NewLocalTransaction("test-tx", nil)

			// 测试并发访问事务状态
			status := tx.GetStatus()
			assert.Equal(t, StatusActive, status)

			// 测试并发获取事务 ID
			txID := tx.GetID()
			assert.Equal(t, "test-tx", txID)

			// 测试并发获取事务类型
			txType := tx.GetType()
			assert.Equal(t, LocalTransaction, txType)
		}(i)
	}

	wg.Wait()
}

func TestTransactionManager_Concurrent(t *testing.T) {
	tm := NewTransactionManager()

	var wg sync.WaitGroup
	goroutines := 20

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			ctx := context.Background()
			
			// 只测试 XA 事务，因为它不需要真实数据库
			tx, err := tm.Begin(ctx, XATransaction)
			assert.NoError(t, err)
			assert.NotNil(t, tx)
			assert.Equal(t, StatusActive, tx.GetStatus())
		}(i)
	}

	wg.Wait()
}

// 基准测试
func BenchmarkLocalTransaction_GetStatus(b *testing.B) {
	tx := NewLocalTransaction("bench-tx", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx.GetStatus()
	}
}

func BenchmarkTransactionManager_BeginXA(b *testing.B) {
	tm := NewTransactionManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		tx, _ := tm.Begin(ctx, XATransaction)
		if tx != nil {
			tx.Rollback(ctx)
		}
	}
}

func BenchmarkGenerateTransactionID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateTransactionID()
	}
}