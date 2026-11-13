package service_repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	gorm_mysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestDebugQueryEmployee(t *testing.T) {
	// 避免在普通自动化测试中一直运行
	if os.Getenv("RUN_DEBUG") != "true" {
		t.Skip("Skipping debug test")
	}

	// 1. 连接数据库
	dsn := "root:Azh501azh@tcp(127.0.0.1:3307)/?charset=utf8mb4&parseTime=true&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("连接失败: %v", err)
	}
	defer db.Close()

	globalDB, err := gorm.Open(gorm_mysql.New(gorm_mysql.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("GORM 初始化失败: %v", err)
	}

	ctx := context.Background()

	// 2. 查询名为 "安贞桦" 的 employee 记录
	var employees []map[string]interface{}
	if err := globalDB.Table("general.employee").WithContext(ctx).Where("name = ?", "安贞桦").Find(&employees).Error; err != nil {
		t.Fatalf("查询 employee 失败: %v", err)
	}
	fmt.Printf("\n=== 安贞桦 Employee 记录 ===\n")
	for _, emp := range employees {
		fmt.Printf("%+v\n", emp)
	}

	// 3. 查询 employee_id 对应的 account 记录
	var accounts []map[string]interface{}
	if len(employees) > 0 {
		empID := employees[0]["id"]
		if err := globalDB.Table("general.account").WithContext(ctx).Where("employee_id = ?", empID).Find(&accounts).Error; err != nil {
			t.Fatalf("查询 account 失败: %v", err)
		}
		fmt.Printf("\n=== 关联 Account 记录 ===\n")
		for _, acc := range accounts {
			fmt.Printf("%+v\n", acc)
		}
	}

	// 4. 查询 employee_role 表中安贞桦的记录
	var roles []map[string]interface{}
	if len(employees) > 0 {
		empID := employees[0]["id"]
		if err := globalDB.Table("quote_manage.employee_role").WithContext(ctx).Where("employee_id = ?", empID).Find(&roles).Error; err != nil {
			t.Fatalf("查询 employee_role 失败: %v", err)
		}
		fmt.Printf("\n=== 关联 EmployeeRole 记录 ===\n")
		for _, r := range roles {
			fmt.Printf("%+v\n", r)
		}
	}

	// 5. 💡 使用真正的 Service 实例查询
	if len(employees) > 0 {
		var empID int32
		idVal := employees[0]["id"]
		switch v := idVal.(type) {
		case int32:
			empID = v
		case uint32:
			empID = int32(v)
		case int64:
			empID = int32(v)
		case int:
			empID = int32(v)
		default:
			empID = 238
		}
		authSvc := NewAuthService(globalDB)
		// 转换成具体的 authService 以便调用内部的 repo
		svcImpl, ok := authSvc.(*authService)
		if !ok {
			t.Fatalf("authSvc is not *authService")
		}
		roleRecord, err := svcImpl.repo.GetEmployeeRoleByEmployeeID(ctx, empID)
		if err != nil {
			fmt.Printf("\n❌ Repository 获取 EmployeeRole 报错: %v\n", err)
		} else if roleRecord == nil {
			fmt.Printf("\n⚠️ Repository 获取 EmployeeRole 结果为 nil\n")
		} else {
			roleStr := "nil"
			if roleRecord.Role != nil {
				roleStr = *roleRecord.Role
			}
			fmt.Printf("\n✅ Repository 获取 EmployeeRole 成功: id=%v, role=%s\n", roleRecord.ID, roleStr)
		}
	}

	fmt.Printf("==============================\n\n")
}

func TestDebugAuthLogin(t *testing.T) {
	if os.Getenv("RUN_DEBUG") != "true" {
		t.Skip("Skipping debug test")
	}

	// 1. 连接数据库
	dsn := "root:Azh501azh@tcp(127.0.0.1:3307)/?charset=utf8mb4&parseTime=true&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("连接失败: %v", err)
	}
	defer db.Close()

	globalDB, err := gorm.Open(gorm_mysql.New(gorm_mysql.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("GORM 初始化失败: %v", err)
	}

	ctx := context.Background()

	// 2. 初始化真正的 service
	authSvc := NewAuthService(globalDB)

	// 3. 执行真实 Login 接口
	username := "S01055"
	password := "202310"
	fmt.Printf("\n🚀 正在测试后端 Service 层 Login 接口，用户: %s ...\n", username)

	loginUser, err := authSvc.Login(ctx, username, password)
	if err != nil {
		t.Fatalf("❌ Login 接口报错: %v", err)
	}

	fmt.Printf("✅ Login 接口成功返回用户数据:\n")
	fmt.Printf("   ID: %d\n", loginUser.ID)
	if loginUser.EmployeeNumber != nil {
		fmt.Printf("   EmployeeNumber: %s\n", *loginUser.EmployeeNumber)
	}
	if loginUser.Name != nil {
		fmt.Printf("   Name: %s\n", *loginUser.Name)
	}
	if loginUser.Department != nil {
		fmt.Printf("   Department: %s\n", *loginUser.Department)
	}
	fmt.Printf("   Role: %s\n\n", loginUser.Role)
}
