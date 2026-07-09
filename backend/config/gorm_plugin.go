package config

import (
	"strings"

	"gorm.io/gorm"
)

// AddDatabasePrefixCallback 为 GORM 的增删改查操作动态拼接数据库名前缀。
// 它是多数据库在同一个 MySQL 实例内共享连接、支持跨库 JOIN 和跨库事务的最优雅解决方案。
func AddDatabasePrefixCallback(db *gorm.DB) {
	prefixFunc := func(tx *gorm.DB) {
		// 1. tx.Statement.Schema 不为 nil 时，说明它是通过 Model struct 进行的操作
		if tx.Statement.Schema != nil {
			modelType := tx.Statement.Schema.ModelType
			pkgPath := modelType.PkgPath()

			// 提取包路径中 "dao/model/" 之后的段作为数据库名
			var dbName string
			const anchor = "/dao/model/"
			if idx := strings.Index(pkgPath, anchor); idx != -1 {
				dbName = pkgPath[idx+len(anchor):]
				// 如果包含子层级，只取第一段，比如 "general"
				if slashIdx := strings.Index(dbName, "/"); slashIdx != -1 {
					dbName = dbName[:slashIdx]
				}
			}

			if dbName != "" {
				// 获取当前的表名（如果有显式指定的 Statement.Table，否则取 Schema.Table）
				tableName := tx.Statement.Table
				if tableName == "" {
					tableName = tx.Statement.Schema.Table
				}

				// 如果表名不为空，且不带有 "."（说明还没拼上数据库名），就拼接上对应的数据库名前缀
				if tableName != "" && !strings.Contains(tableName, ".") {
					tx.Statement.Table = dbName + "." + tableName
				}
			}
		}
	}

	// 注册在 GORM 内建的各个 Callback 前，使得在编译生成具体 SQL 之前生效
	_ = db.Callback().Create().Before("gorm:create").Register("db_prefix:before_create", prefixFunc)
	_ = db.Callback().Query().Before("gorm:query").Register("db_prefix:before_query", prefixFunc)
	_ = db.Callback().Update().Before("gorm:update").Register("db_prefix:before_update", prefixFunc)
	_ = db.Callback().Delete().Before("gorm:delete").Register("db_prefix:before_delete", prefixFunc)
}
