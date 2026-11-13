package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"
	_ "time/tzdata"

	general_handler "csun_server-backend/general/router_handler"
	general_service "csun_server-backend/general/service_repository"
	hr_handler "csun_server-backend/hr_sync/router_handler"
	hr_service "csun_server-backend/hr_sync/service_repository"
	quote_handler "csun_server-backend/quote_manage/router_handler"
	quote_service "csun_server-backend/quote_manage/service_repository"

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

// initDB 从环境变量读取连接参数并建立 MySQL 连接，不绑定to特定数据库
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

	// 增加数据库连接重试逻辑，等待 SSH 隧道建立就绪
	maxRetries := 15
	var pingErr error
	for i := 1; i <= maxRetries; i++ {
		pingErr = d.Ping()
		if pingErr == nil {
			if i > 1 {
				log.Printf("✅ 数据库/SSH 隧道已就绪，连接成功！")
			}
			return d
		}
		log.Printf("⏳ 等待数据库/SSH 隧道就绪中 (%d/%d)...", i, maxRetries)
		time.Sleep(1 * time.Second)
	}

	log.Fatalf("❌ 数据库 Ping 失败 (超过最大重试次数): %v", pingErr)
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

	dbEngine := &general_service.DBEngine{
		General:     globalDB,
		Product:     globalDB,
		QuoteManage: globalDB,
	}

	// ==================== 通用基础模块 (general) ====================
	// 初始化 Auth 依赖链
	authService := general_service.NewAuthService(dbEngine.General)
	authHandler := general_handler.NewAuthHandler(authService)

	// 初始化 Password 依赖链
	passwordService := general_service.NewPasswordService(dbEngine.General)
	passwordHandler := general_handler.NewPasswordHandler(passwordService)

	// 初始化 DataTable 依赖链
	dataTableService := general_service.NewDataTableService(dbEngine)
	dataTableHandler := general_handler.NewDataTableHandler(dataTableService)

	// ==================== 报价管理应用 (quote_manage) ====================
	// 初始化 CreateQuote 依赖链
	createQuoteService := quote_service.NewCreateQuoteService(dbEngine)
	createQuoteHandler := quote_handler.NewCreateQuoteHandler(createQuoteService)

	// 初始化 MyApproveQuery 依赖链
	myApproveQueryService := quote_service.NewMyApproveQueryService(dbEngine)
	myApproveQueryHandler := quote_handler.NewMyApproveQueryHandler(myApproveQueryService)

	// 初始化 FilingLook 依赖链
	filingLookService := quote_service.NewFilingLookService(dbEngine)
	filingLookHandler := quote_handler.NewFilingLookHandler(filingLookService)

	// 初始化 MyApplyQuery 依赖链
	myApplyQueryService := quote_service.NewMyApplyQueryService(dbEngine)
	myApplyQueryHandler := quote_handler.NewMyApplyQueryHandler(myApplyQueryService)

	// 初始化 ApproveHandle 依赖链
	approveHandleService := quote_service.NewApproveHandleService(dbEngine)
	approveHandleHandler := quote_handler.NewApproveHandleHandler(approveHandleService)

	// 初始化 WithdrawQuote 依赖链
	withdrawQuoteService := quote_service.NewWithdrawQuoteService(dbEngine)
	withdrawQuoteHandler := quote_handler.NewWithdrawQuoteHandler(withdrawQuoteService)

	// ==================== 人事同步应用 (hr_sync) ====================
	// 初始化 Callback 依赖链
	hrCallbackService := hr_service.NewCallbackService(dbEngine)
	hrCallbackHandler := hr_handler.NewCallbackHandler(hrCallbackService)

	// gin.Default() 创建一个带 Logger 与 Recovery 中间件的引擎
	r := gin.Default()

	// 注册通用模块 (general) 路由
	general_handler.RegisterAuthRoutes(r, authHandler)
	general_handler.RegisterPasswordRoutes(r, passwordHandler)
	general_handler.RegisterDataTableRoutes(r, dataTableHandler)
	general_handler.RegisterCOSRoutes(r)

	// 注册报价管理应用 (quote_manage) 路由
	quote_handler.RegisterCreateQuoteRoutes(r, createQuoteHandler)
	quote_handler.RegisterMyApproveQueryRoutes(r, myApproveQueryHandler)
	quote_handler.RegisterFilingLookRoutes(r, filingLookHandler)
	quote_handler.RegisterMyApplyQueryRoutes(r, myApplyQueryHandler)
	quote_handler.RegisterApproveHandleRoutes(r, approveHandleHandler)
	quote_handler.RegisterWithdrawQuoteRoutes(r, withdrawQuoteHandler)

	// 注册人事同步应用 (hr_sync) 路由
	hr_handler.RegisterCallbackRoutes(r, hrCallbackHandler)

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
