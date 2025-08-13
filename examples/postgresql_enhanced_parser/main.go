package main

import (
	"encoding/json"
	"fmt"
	"go-sharding/pkg/parser"
	"log"
	"strings"
)

func main() {
	fmt.Println("=== PostgreSQL 增强解析器演示 ===")
	fmt.Println()

	// 创建增强的 PostgreSQL 解析器
	enhancedParser := parser.NewPostgreSQLEnhancedParser()

	// 测试用例
	testCases := []struct {
		name string
		sql  string
	}{
		{
			name: "简单 SELECT 查询",
			sql:  "SELECT id, name, email FROM users WHERE age > 18 ORDER BY name LIMIT 10",
		},
		{
			name: "复杂 JOIN 查询",
			sql: `SELECT u.id, u.name, p.title, c.content 
				   FROM users u 
				   INNER JOIN posts p ON u.id = p.user_id 
				   LEFT JOIN comments c ON p.id = c.post_id 
				   WHERE u.active = true AND p.published_at > '2023-01-01'
				   ORDER BY p.published_at DESC`,
		},
		{
			name: "带子查询的查询",
			sql: `SELECT u.name, 
						 (SELECT COUNT(*) FROM posts WHERE user_id = u.id) as post_count,
						 (SELECT AVG(rating) FROM reviews WHERE user_id = u.id) as avg_rating
				   FROM users u 
				   WHERE u.id IN (SELECT DISTINCT user_id FROM posts WHERE published = true)`,
		},
		{
			name: "CTE (Common Table Expression) 查询",
			sql: `WITH active_users AS (
						 SELECT id, name FROM users WHERE active = true
				   ),
				   user_stats AS (
						 SELECT user_id, COUNT(*) as post_count 
						 FROM posts 
						 GROUP BY user_id
				   )
				   SELECT au.name, us.post_count
				   FROM active_users au
				   JOIN user_stats us ON au.id = us.user_id
				   ORDER BY us.post_count DESC`,
		},
		{
			name: "窗口函数查询",
			sql: `SELECT 
						 name,
						 salary,
						 department,
						 ROW_NUMBER() OVER (PARTITION BY department ORDER BY salary DESC) as rank,
						 AVG(salary) OVER (PARTITION BY department) as dept_avg_salary,
						 LAG(salary, 1) OVER (ORDER BY salary) as prev_salary
				   FROM employees`,
		},
		{
			name: "INSERT with RETURNING",
			sql:  "INSERT INTO users (name, email, age) VALUES ('John Doe', 'john@example.com', 30) RETURNING id, created_at",
		},
		{
			name: "UPDATE with 子查询",
			sql: `UPDATE posts 
				   SET view_count = view_count + 1,
					   last_viewed = NOW()
				   WHERE id IN (SELECT post_id FROM user_favorites WHERE user_id = 123)
				   RETURNING id, view_count`,
		},
		{
			name: "复杂的递归 CTE",
			sql: `WITH RECURSIVE employee_hierarchy AS (
						 SELECT id, name, manager_id, 1 as level
						 FROM employees 
						 WHERE manager_id IS NULL
						 UNION ALL
						 SELECT e.id, e.name, e.manager_id, eh.level + 1
						 FROM employees e
						 JOIN employee_hierarchy eh ON e.manager_id = eh.id
				   )
				   SELECT * FROM employee_hierarchy ORDER BY level, name`,
		},
		{
			name: "PostgreSQL 特定功能",
			sql: `SELECT 
						 id,
						 name,
						 tags::jsonb,
						 ARRAY_AGG(category) as categories,
						 STRING_AGG(description, '; ') as descriptions,
						 coordinates::point,
						 metadata @> '{"featured": true}' as is_featured
				   FROM products p
				   WHERE tags ?& array['electronics', 'gadgets']
				   GROUP BY id, name, tags, coordinates, metadata`,
		},
	}

	// 测试每个 SQL 语句
	for i, testCase := range testCases {
		fmt.Printf("\n%d. %s\n", i+1, testCase.name)
		fmt.Println(strings.Repeat("=", 50))
		fmt.Printf("SQL: %s\n\n", testCase.sql)

		// 分析 SQL
		analysis, err := enhancedParser.AnalyzeSQL(testCase.sql)
		if err != nil {
			fmt.Printf("❌ 分析失败: %v\n", err)
			continue
		}

		// 显示基本信息
		fmt.Printf("📊 基本信息:\n")
		fmt.Printf("   类型: %s\n", analysis.Type)
		fmt.Printf("   表: %v\n", analysis.Tables)
		fmt.Printf("   复杂度分数: %d\n", analysis.Complexity.Score)

		// 显示 CTE 信息
		if len(analysis.CTEs) > 0 {
			fmt.Printf("\n🔗 CTE 信息:\n")
			for _, cte := range analysis.CTEs {
				fmt.Printf("   - %s (递归: %v, 表: %v)\n", cte.Name, cte.Recursive, cte.Tables)
			}
		}

		// 显示 JOIN 信息
		if len(analysis.Joins) > 0 {
			fmt.Printf("\n🔀 JOIN 信息:\n")
			for _, join := range analysis.Joins {
				fmt.Printf("   - %s: %s ⟷ %s\n", join.Type, join.LeftTable, join.RightTable)
				if join.Condition != "" {
					fmt.Printf("     条件: %s\n", join.Condition)
				}
			}
		}

		// 显示子查询信息
		if len(analysis.Subqueries) > 0 {
			fmt.Printf("\n🔍 子查询信息:\n")
			for _, subquery := range analysis.Subqueries {
				fmt.Printf("   - 类型: %s, 表: %v\n", subquery.Type, subquery.Tables)
			}
		}

		// 显示窗口函数信息
		if len(analysis.WindowFunctions) > 0 {
			fmt.Printf("\n🪟 窗口函数信息:\n")
			for _, wf := range analysis.WindowFunctions {
				fmt.Printf("   - 函数: %s\n", wf.Function)
				if len(wf.PartitionBy) > 0 {
					fmt.Printf("     PARTITION BY: %v\n", wf.PartitionBy)
				}
				if len(wf.OrderBy) > 0 {
					fmt.Printf("     ORDER BY: %v\n", wf.OrderBy)
				}
			}
		}

		// 显示复杂度指标
		fmt.Printf("\n📈 复杂度指标:\n")
		fmt.Printf("   表数量: %d\n", analysis.Complexity.TableCount)
		fmt.Printf("   JOIN 数量: %d\n", analysis.Complexity.JoinCount)
		fmt.Printf("   子查询数量: %d\n", analysis.Complexity.SubqueryCount)
		fmt.Printf("   CTE 数量: %d\n", analysis.Complexity.CTECount)
		fmt.Printf("   窗口函数数量: %d\n", analysis.Complexity.WindowFuncCount)
		fmt.Printf("   嵌套级别: %d\n", analysis.Complexity.NestingLevel)

		// 显示优化建议
		if len(analysis.Optimizations) > 0 {
			fmt.Printf("\n💡 优化建议:\n")
			for _, opt := range analysis.Optimizations {
				severityIcon := "ℹ️"
				if opt.Severity == "warning" {
					severityIcon = "⚠️"
				} else if opt.Severity == "error" {
					severityIcon = "❌"
				}
				fmt.Printf("   %s [%s] %s\n", severityIcon, opt.Type, opt.Message)
				fmt.Printf("      建议: %s\n", opt.Suggestion)
			}
		}

		// 显示 PostgreSQL 特性
		if len(analysis.PostgreSQLFeatures) > 0 {
			fmt.Printf("\n🐘 PostgreSQL 特性:\n")
			for feature, value := range analysis.PostgreSQLFeatures {
				fmt.Printf("   - %s: %v\n", feature, value)
			}
		}
	}

	fmt.Println("\n=== 增强功能演示 ===")
	fmt.Println()

	// 演示增强的表名提取
	complexSQL := `WITH user_stats AS (
		SELECT user_id, COUNT(*) as post_count FROM posts GROUP BY user_id
	)
	SELECT u.name, us.post_count, 
		   (SELECT COUNT(*) FROM comments WHERE user_id = u.id) as comment_count
	FROM users u
	JOIN user_stats us ON u.id = us.user_id
	WHERE u.id IN (SELECT user_id FROM subscriptions WHERE active = true)`

	fmt.Println("1. 增强的表名提取")
	fmt.Println("SQL:", complexSQL)
	tables, err := enhancedParser.ExtractTablesEnhanced(complexSQL)
	if err != nil {
		log.Printf("表名提取失败: %v", err)
	} else {
		fmt.Println("提取的表名:")
		for category, tableList := range tables {
			fmt.Printf("  %s: %v\n", category, tableList)
		}
	}

	fmt.Println("\n2. SQL 重写演示")
	shardingRules := map[string]string{
		"users":    "users_shard_1",
		"posts":    "posts_shard_1",
		"comments": "comments_shard_1",
	}
	rewrittenSQL, err := enhancedParser.RewriteForSharding(complexSQL, shardingRules)
	if err != nil {
		log.Printf("SQL 重写失败: %v", err)
	} else {
		fmt.Println("重写后的 SQL:")
		fmt.Println(rewrittenSQL)
	}

	fmt.Println("\n3. 复杂 SQL 验证")
	validationErrors, err := enhancedParser.ValidateComplexSQL(complexSQL)
	if err != nil {
		log.Printf("验证失败: %v", err)
	} else if len(validationErrors) > 0 {
		fmt.Println("验证错误:")
		for _, verr := range validationErrors {
			fmt.Printf("  - [%s] %s\n", verr.Type, verr.Message)
		}
	} else {
		fmt.Println("✅ SQL 验证通过")
	}

	fmt.Println("\n4. 表依赖关系分析")
	dependencies, err := enhancedParser.AnalyzeTableDependencies(complexSQL)
	if err != nil {
		log.Printf("依赖分析失败: %v", err)
	} else {
		fmt.Println("表依赖关系:")
		for table, deps := range dependencies {
			fmt.Printf("  %s -> %v\n", table, deps)
		}
	}

	fmt.Println("\n5. 优化建议")
	suggestions, err := enhancedParser.GetOptimizationSuggestions(complexSQL)
	if err != nil {
		log.Printf("优化建议获取失败: %v", err)
	} else {
		fmt.Println("优化建议:")
		for _, suggestion := range suggestions {
			fmt.Printf("  - [%s] %s: %s\n", suggestion.Severity, suggestion.Type, suggestion.Message)
			fmt.Printf("    建议: %s\n", suggestion.Suggestion)
		}
	}

	fmt.Println("\n=== JSON 输出演示 ===")
	analysis, err := enhancedParser.AnalyzeSQL(complexSQL)
	if err != nil {
		log.Printf("分析失败: %v", err)
	} else {
		jsonData, err := json.MarshalIndent(analysis, "", "  ")
		if err != nil {
			log.Printf("JSON 序列化失败: %v", err)
		} else {
			fmt.Println("完整分析结果 (JSON):")
			fmt.Println(string(jsonData))
		}
	}

	fmt.Println("\n🎉 PostgreSQL 增强解析器演示完成!")
}