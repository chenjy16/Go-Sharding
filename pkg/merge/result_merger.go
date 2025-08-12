package merge

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// ResultMerger 结果合并器
type ResultMerger struct {
}

// NewResultMerger 创建结果合并器
func NewResultMerger() *ResultMerger {
	return &ResultMerger{}
}

// MergeContext 合并上下文
type MergeContext struct {
	SQL         string
	OrderByColumns []OrderByColumn
	LimitOffset    int
	LimitCount     int
	GroupByColumns []string
}

// OrderByColumn 排序列
type OrderByColumn struct {
	Column string
	Desc   bool
}

// MergedRows 合并后的结果集
type MergedRows struct {
	columns []string
	rows    [][]interface{}
	index   int
}

// NewMergedRows 创建合并结果集
func NewMergedRows(columns []string, rows [][]interface{}) *MergedRows {
	return &MergedRows{
		columns: columns,
		rows:    rows,
		index:   -1,
	}
}

// Columns 返回列名
func (r *MergedRows) Columns() []string {
	return r.columns
}

// Close 关闭结果集
func (r *MergedRows) Close() error {
	return nil
}

// Next 移动到下一行
func (r *MergedRows) Next(dest []driver.Value) error {
	r.index++
	if r.index >= len(r.rows) {
		return io.EOF
	}
	
	row := r.rows[r.index]
	for i, value := range row {
		if i < len(dest) {
			dest[i] = value
		}
	}
	
	return nil
}

// Merge 合并多个结果集
func (m *ResultMerger) Merge(results []*sql.Rows, ctx *MergeContext) (*MergedRows, error) {
	if len(results) == 0 {
		return NewMergedRows([]string{}, [][]interface{}{}), nil
	}

	// 收集所有行数据
	var allRows [][]interface{}
	var columns []string
	
	for i, rows := range results {
		if i == 0 {
			var err error
			columns, err = rows.Columns()
			if err != nil {
				return nil, fmt.Errorf("failed to get columns: %w", err)
			}
		}
		
		rowData, err := m.scanAllRows(rows, len(columns))
		if err != nil {
			return nil, fmt.Errorf("failed to scan rows: %w", err)
		}
		
		allRows = append(allRows, rowData...)
	}

	// 应用合并逻辑
	mergedRows := allRows
	
	// 排序
	if len(ctx.OrderByColumns) > 0 {
		mergedRows = m.sortRows(mergedRows, columns, ctx.OrderByColumns)
	}
	
	// 分组聚合
	if len(ctx.GroupByColumns) > 0 {
		mergedRows = m.groupRows(mergedRows, columns, ctx.GroupByColumns)
	}
	
	// 限制结果
	if ctx.LimitCount > 0 {
		start := ctx.LimitOffset
		end := start + ctx.LimitCount
		if start < len(mergedRows) {
			if end > len(mergedRows) {
				end = len(mergedRows)
			}
			mergedRows = mergedRows[start:end]
		} else {
			mergedRows = [][]interface{}{}
		}
	}

	return NewMergedRows(columns, mergedRows), nil
}

// scanAllRows 扫描所有行数据
func (m *ResultMerger) scanAllRows(rows *sql.Rows, columnCount int) ([][]interface{}, error) {
	var allRows [][]interface{}
	
	for rows.Next() {
		row := make([]interface{}, columnCount)
		scanArgs := make([]interface{}, columnCount)
		for i := range row {
			scanArgs[i] = &row[i]
		}
		
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		
		allRows = append(allRows, row)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	
	return allRows, nil
}

// sortRows 排序行数据
func (m *ResultMerger) sortRows(rows [][]interface{}, columns []string, orderBy []OrderByColumn) [][]interface{} {
	// 创建列名到索引的映射
	columnIndex := make(map[string]int)
	for i, col := range columns {
		columnIndex[col] = i
	}
	
	sort.Slice(rows, func(i, j int) bool {
		for _, orderCol := range orderBy {
			colIdx, exists := columnIndex[orderCol.Column]
			if !exists {
				continue
			}
			
			val1 := rows[i][colIdx]
			val2 := rows[j][colIdx]
			
			cmp := m.compareValues(val1, val2)
			if cmp != 0 {
				if orderCol.Desc {
					return cmp > 0
				}
				return cmp < 0
			}
		}
		return false
	})
	
	return rows
}

// groupRows 分组聚合行数据
func (m *ResultMerger) groupRows(rows [][]interface{}, columns []string, groupBy []string) [][]interface{} {
	if len(rows) == 0 {
		return rows
	}
	
	// 创建列名到索引的映射
	columnIndex := make(map[string]int)
	for i, col := range columns {
		columnIndex[col] = i
	}
	
	// 按分组键分组
	groups := make(map[string][]int)
	for i, row := range rows {
		key := m.buildGroupKey(row, columns, groupBy, columnIndex)
		groups[key] = append(groups[key], i)
	}
	
	// 聚合每个分组
	var result [][]interface{}
	for _, indices := range groups {
		if len(indices) > 0 {
			aggregatedRow := m.aggregateGroup(rows, indices, columns)
			result = append(result, aggregatedRow)
		}
	}
	
	return result
}

// buildGroupKey 构建分组键
func (m *ResultMerger) buildGroupKey(row []interface{}, columns []string, groupBy []string, columnIndex map[string]int) string {
	var keyParts []string
	
	for _, col := range groupBy {
		if idx, exists := columnIndex[col]; exists && idx < len(row) {
			keyParts = append(keyParts, fmt.Sprintf("%v", row[idx]))
		}
	}
	
	return strings.Join(keyParts, "|")
}

// aggregateGroup 聚合分组数据
func (m *ResultMerger) aggregateGroup(rows [][]interface{}, indices []int, columns []string) []interface{} {
	if len(indices) == 0 {
		return nil
	}
	
	// 使用第一行作为基础
	result := make([]interface{}, len(columns))
	copy(result, rows[indices[0]])
	
	// 对于聚合函数列，需要重新计算
	// 这里简化处理，实际应该解析 SQL 中的聚合函数
	for i, col := range columns {
		if m.isAggregateColumn(col) {
			result[i] = m.calculateAggregate(col, rows, indices, i)
		}
	}
	
	return result
}

// isAggregateColumn 检查是否为聚合列
func (m *ResultMerger) isAggregateColumn(column string) bool {
	upper := strings.ToUpper(column)
	return strings.Contains(upper, "COUNT(") ||
		   strings.Contains(upper, "SUM(") ||
		   strings.Contains(upper, "AVG(") ||
		   strings.Contains(upper, "MIN(") ||
		   strings.Contains(upper, "MAX(")
}

// calculateAggregate 计算聚合值
func (m *ResultMerger) calculateAggregate(column string, rows [][]interface{}, indices []int, columnIndex int) interface{} {
	upper := strings.ToUpper(column)
	
	if strings.Contains(upper, "COUNT(") {
		return len(indices)
	}
	
	if strings.Contains(upper, "SUM(") {
		var sum float64
		for _, idx := range indices {
			if val := m.toFloat64(rows[idx][columnIndex]); val != nil {
				sum += *val
			}
		}
		return sum
	}
	
	if strings.Contains(upper, "AVG(") {
		var sum float64
		var count int
		for _, idx := range indices {
			if val := m.toFloat64(rows[idx][columnIndex]); val != nil {
				sum += *val
				count++
			}
		}
		if count > 0 {
			return sum / float64(count)
		}
		return nil
	}
	
	if strings.Contains(upper, "MIN(") {
		var min interface{}
		for _, idx := range indices {
			val := rows[idx][columnIndex]
			if min == nil || m.compareValues(val, min) < 0 {
				min = val
			}
		}
		return min
	}
	
	if strings.Contains(upper, "MAX(") {
		var max interface{}
		for _, idx := range indices {
			val := rows[idx][columnIndex]
			if max == nil || m.compareValues(val, max) > 0 {
				max = val
			}
		}
		return max
	}
	
	// 默认返回第一个值
	if len(indices) > 0 {
		return rows[indices[0]][columnIndex]
	}
	return nil
}

// compareValues 比较两个值
func (m *ResultMerger) compareValues(a, b interface{}) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}
	
	// 尝试转换为数字比较
	if numA := m.toFloat64(a); numA != nil {
		if numB := m.toFloat64(b); numB != nil {
			if *numA < *numB {
				return -1
			} else if *numA > *numB {
				return 1
			}
			return 0
		}
	}
	
	// 字符串比较
	strA := fmt.Sprintf("%v", a)
	strB := fmt.Sprintf("%v", b)
	if strA < strB {
		return -1
	} else if strA > strB {
		return 1
	}
	return 0
}

// toFloat64 尝试转换为 float64
func (m *ResultMerger) toFloat64(value interface{}) *float64 {
	switch v := value.(type) {
	case int:
		f := float64(v)
		return &f
	case int32:
		f := float64(v)
		return &f
	case int64:
		f := float64(v)
		return &f
	case float32:
		f := float64(v)
		return &f
	case float64:
		return &v
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return &f
		}
	}
	return nil
}