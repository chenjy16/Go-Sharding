package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-sharding/pkg/transaction"
)

func main() {
	// 创建事务管理器
	tm := transaction.NewTransactionManager()
	defer tm.Close()

	// 开始BASE事务
	ctx := context.Background()
	tx, err := tm.Begin(ctx, transaction.BaseTransaction)
	if err != nil {
		log.Fatalf("Failed to begin BASE transaction: %v", err)
	}

	// 类型断言为BASE事务
	baseTx, ok := tx.(*transaction.BASETransactionImpl)
	if !ok {
		log.Fatalf("Expected BASETransactionImpl, got %T", tx)
	}

	fmt.Printf("Created BASE transaction: %s\n", baseTx.GetID())
	fmt.Printf("Transaction type: %v\n", baseTx.GetType())
	fmt.Printf("Initial status: %v\n", baseTx.GetStatus())

	// 添加业务操作
	op1 := transaction.BASEOperation{
		Type:       "INSERT",
		SQL:        "INSERT INTO orders (user_id, amount) VALUES (?, ?)",
		DataSource: "order_db",
		Parameters: []interface{}{123, 99.99},
	}

	op2 := transaction.BASEOperation{
		Type:       "UPDATE",
		SQL:        "UPDATE inventory SET quantity = quantity - ? WHERE product_id = ?",
		DataSource: "inventory_db",
		Parameters: []interface{}{1, 456},
	}

	// 添加操作到事务
	if err := baseTx.AddOperation(op1); err != nil {
		log.Fatalf("Failed to add operation 1: %v", err)
	}

	if err := baseTx.AddOperation(op2); err != nil {
		log.Fatalf("Failed to add operation 2: %v", err)
	}

	// 添加补偿操作
	comp1 := transaction.BASECompensation{
		OperationID: "op1",
		SQL:         "DELETE FROM orders WHERE user_id = ? AND amount = ?",
		DataSource:  "order_db",
		Parameters:  []interface{}{123, 99.99},
	}

	comp2 := transaction.BASECompensation{
		OperationID: "op2",
		SQL:         "UPDATE inventory SET quantity = quantity + ? WHERE product_id = ?",
		DataSource:  "inventory_db",
		Parameters:  []interface{}{1, 456},
	}

	if err := baseTx.AddCompensation(comp1); err != nil {
		log.Fatalf("Failed to add compensation 1: %v", err)
	}

	if err := baseTx.AddCompensation(comp2); err != nil {
		log.Fatalf("Failed to add compensation 2: %v", err)
	}

	fmt.Printf("Added %d operations and %d compensations\n", 
		len(baseTx.GetOperations()), len(baseTx.GetCompensations()))

	// 提交事务
	fmt.Println("Committing transaction...")
	if err := baseTx.Commit(ctx); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	// 等待异步执行完成
	time.Sleep(200 * time.Millisecond)

	// 检查最终状态
	finalStatus := baseTx.GetStatus()
	fmt.Printf("Final transaction status: %v\n", finalStatus)

	// 显示操作状态
	operations := baseTx.GetOperations()
	for i, op := range operations {
		fmt.Printf("Operation %d: Type=%s, Status=%s\n", i+1, op.Type, op.Status)
	}

	// 演示超时功能
	fmt.Println("\n--- Timeout Example ---")
	timeoutTx := transaction.NewBASETransaction("timeout-tx")
	timeoutTx.SetTimeout(1 * time.Millisecond)
	
	fmt.Printf("Transaction created at: %v\n", time.Now())
	time.Sleep(10 * time.Millisecond)
	
	if timeoutTx.IsExpired() {
		fmt.Println("Transaction has expired")
	} else {
		fmt.Println("Transaction is still active")
	}

	// 演示回滚功能
	fmt.Println("\n--- Rollback Example ---")
	rollbackTx := transaction.NewBASETransaction("rollback-tx")
	
	// 添加补偿操作
	rollbackComp := transaction.BASECompensation{
		OperationID: "failed-op",
		SQL:         "ROLLBACK OPERATION",
		DataSource:  "test_db",
	}
	
	if err := rollbackTx.AddCompensation(rollbackComp); err != nil {
		log.Fatalf("Failed to add rollback compensation: %v", err)
	}
	
	fmt.Printf("Rollback transaction status before rollback: %v\n", rollbackTx.GetStatus())
	
	if err := rollbackTx.Rollback(ctx); err != nil {
		log.Fatalf("Failed to rollback transaction: %v", err)
	}
	
	// 等待异步执行完成
	time.Sleep(100 * time.Millisecond)
	
	fmt.Printf("Rollback transaction status after rollback: %v\n", rollbackTx.GetStatus())

	fmt.Println("\nBASE transaction example completed successfully!")
}