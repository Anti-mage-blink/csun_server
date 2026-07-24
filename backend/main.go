package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"
	_ "time/tzdata"

	"csun_server-backend/router_handler"
	"csun_server-backend/service_repository"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	gorm_mysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	var err error
	time.Local, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Printf("初始化东八区时区失败: %v", err)
	}
}

// initDB 从环境变量读取连接参数并建立 MySQL 连接，不绑定 to 特定数据库
func initDB() *sql.DB {
	// 若未设置环境变量，则使用默认值
	if os.Getenv("DB_HOST") == "" {
		os.Setenv("DB_HOST", "127.0.0.1")
	}
	if os.Getenv("DB_PORT") == "" {
		os.Setenv("DB_PORT", "3307")
	}
	if os.Getenv("DB_USER") == "" {
		os.Setenv("DB_USER", "root")
	}
	if os.Getenv("DB_PASSWORD") == "" {
		os.Setenv("DB_PASSWORD", "Azh501azh")
	}

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

	dbEngine := &service_repository.DBEngine{
		General:     globalDB,
		Product:     globalDB,
		QuoteManage: globalDB,
	}

	// 初始化 Auth 依赖链
	authService := service_repository.NewAuthService(dbEngine.General)
	authHandler := router_handler.NewAuthHandler(authService)

	// 初始化 Password 依赖链
	passwordService := service_repository.NewPasswordService(dbEngine.General)
	passwordHandler := router_handler.NewPasswordHandler(passwordService)

	// 初始化 CreateQuote 依赖链
	createQuoteService := service_repository.NewCreateQuoteService(dbEngine)
	createQuoteHandler := router_handler.NewCreateQuoteHandler(createQuoteService)

	// 初始化 QueryNeedApprove 依赖链
	queryNeedApproveService := service_repository.NewQueryNeedApproveService(dbEngine)
	queryNeedApproveHandler := router_handler.NewQueryNeedApproveHandler(queryNeedApproveService)

	// 初始化 FilingLook 依赖链
	filingLookService := service_repository.NewFilingLookService(dbEngine)
	filingLookHandler := router_handler.NewFilingLookHandler(filingLookService)

	// 初始化 MyApplyQuery 依赖链
	myApplyQueryService := service_repository.NewMyApplyQueryService(dbEngine)
	myApplyQueryHandler := router_handler.NewMyApplyQueryHandler(myApplyQueryService)

	// 初始化 ApproveHandle 依赖链
	approveHandleService := service_repository.NewApproveHandleService(dbEngine)
	approveHandleHandler := router_handler.NewApproveHandleHandler(approveHandleService)

	// gin.Default() 创建一个带 Logger 与 Recovery 中间件的引擎
	r := gin.Default()

	// 注册 Auth 路由
	router_handler.RegisterAuthRoutes(r, authHandler)
	// 注册 Password 路由
	router_handler.RegisterPasswordRoutes(r, passwordHandler)
	// 注册 CreateQuote 路由
	router_handler.RegisterCreateQuoteRoutes(r, createQuoteHandler)
	// 注册 QueryNeedApprove 路由
	router_handler.RegisterQueryNeedApproveRoutes(r, queryNeedApproveHandler)
	// 注册 FilingLook 路由
	router_handler.RegisterFilingLookRoutes(r, filingLookHandler)
	// 注册 MyApplyQuery 路由
	router_handler.RegisterMyApplyQueryRoutes(r, myApplyQueryHandler)
	// 注册 ApproveHandle 路由
	router_handler.RegisterApproveHandleRoutes(r, approveHandleHandler)

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
