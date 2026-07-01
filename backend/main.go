package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	gorm_mysql "gorm.io/driver/mysql"
	"gorm.io/gorm"

	"csun_server-backend/service"
)

// initDB 从环境变量读取连接参数并建立 MySQL 连接
func initDB() *sql.DB {
	dsn := os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") +
		"@tcp(" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + ")/" +
		os.Getenv("DB_NAME") + "?charset=utf8mb4&parseTime=true&loc=Local"
	d, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("sql.Open 失败: %v", err)
	}
	if err := d.Ping(); err != nil {
		log.Fatalf("数据库 Ping 失败: %v", err)
	}
	return d
}

func main() {
	db := initDB()
	defer db.Close()

	// 初始化 GORM，复用已有的 sql.DB
	gormDB, err := gorm.Open(gorm_mysql.New(gorm_mysql.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		log.Fatalf("GORM 初始化失败: %v", err)
	}

	// gin.Default() 创建一个带 Logger 与 Recovery 中间件的引擎
	r := gin.Default()

	// 根路由：返回 JSON 格式的 Hello World
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":   "Hello, World!",
			"framework": "Gin v1.12.0",
			"go":        "1.26",
		})
	})

	// 健康检查示例路由
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// 获取新报价单号接口
	r.GET("/quote-manage/new-code", func(c *gin.Context) {
		code, err := service.GetNewQuoteCode(c.Request.Context(), gormDB)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "生成报价单编号失败",
				"detail":  err.Error(),
				"success": false,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    code,
			"success": true,
		})
	})

	// 查询 quote_manage 数据库的所有数据表名
	r.GET("/tables", func(c *gin.Context) {
		rows, err := db.Query("SHOW TABLES")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "查询表失败",
				"detail":  err.Error(),
				"success": false,
			})
			return
		}
		defer rows.Close()

		var tables []string
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "读取表名失败",
					"detail":  err.Error(),
					"success": false,
				})
				return
			}
			tables = append(tables, name)
		}

		// 确保 tables 为 nil 时返回空数组而非 null
		if tables == nil {
			tables = []string{}
		}

		c.JSON(http.StatusOK, gin.H{
			"database": os.Getenv("DB_NAME"),
			"tables":   tables,
			"count":    len(tables),
			"success":  true,
		})
	})

	// 默认监听 0.0.0.0:8080
	log.Printf("Gin server starting on http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
