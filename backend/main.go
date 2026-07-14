package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"csun_server-backend/handler"
	"csun_server-backend/repository"
	"csun_server-backend/router"
	"csun_server-backend/service"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	gorm_mysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// initDB 从环境变量读取连接参数并建立 MySQL 连接，不绑定到特定数据库
func initDB() *sql.DB {
	dsn := os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") +
		"@tcp(" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + ")/" +
		"?charset=utf8mb4&parseTime=true&loc=Local"
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

	// 初始化唯一的 GORM 实例
	globalDB, err := gorm.Open(gorm_mysql.New(gorm_mysql.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		log.Fatalf("GORM 初始化失败: %v", err)
	}

	dbEngine := &repository.DBEngine{
		General:     globalDB,
		Product:     globalDB,
		QuoteManage: globalDB,
	}

	// 初始化 Auth 依赖链
	authRepo := repository.NewAuthRepository(dbEngine.General)
	authService := service.NewAuthService(authRepo)
	authHandler := handler.NewAuthHandler(authService)

	// 初始化 Password 依赖链
	passwordRepo := repository.NewPasswordRepository(dbEngine.General)
	passwordService := service.NewPasswordService(passwordRepo)
	passwordHandler := handler.NewPasswordHandler(passwordService)

	// 初始化 CreateQuote 依赖链
	createQuoteRepo := repository.NewCreateQuoteRepository(dbEngine)
	createQuoteService := service.NewCreateQuoteService(createQuoteRepo)
	createQuoteHandler := handler.NewCreateQuoteHandler(createQuoteService)

	// gin.Default() 创建一个带 Logger 与 Recovery 中间件的引擎
	r := gin.Default()

	// 注册 Auth 路由
	router.RegisterAuthRoutes(r, authHandler)
	// 注册 Password 路由
	router.RegisterPasswordRoutes(r, passwordHandler)
	// 注册 CreateQuote 路由
	router.RegisterCreateQuoteRoutes(r, createQuoteHandler)

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

	// 默认监听 0.0.0.0:8080
	log.Printf("Gin server starting on http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
