package repository

import "gorm.io/gorm"

// DBEngine 整合了系统中的三个核心业务数据库连接
type DBEngine struct {
	General     *gorm.DB
	Product     *gorm.DB
	QuoteManage *gorm.DB
}
