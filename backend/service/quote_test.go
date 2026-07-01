package service

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	gorm_mysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestGetNewQuoteCode(t *testing.T) {
	// 1. 初始化 sqlmock 和 gorm.DB
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建 sqlmock 失败: %v", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open(gorm_mysql.New(gorm_mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true, // 必须跳过版本查询，以便完美匹配 mock
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("初始化 GORM 失败: %v", err)
	}

	todayStr := time.Now().Format("20060102")
	prefix := "BJ-" + todayStr

	t.Run("数据库为空，应该返回 -001", func(t *testing.T) {
		mock.ExpectQuery(`SELECT \* FROM ` + "`quote`" + ` WHERE ` + "`quote`" + `\.` + "`quote_code`" + ` LIKE \? ORDER BY ` + "`quote`" + `\.` + "`quote_code`" + ` DESC,` + "`quote`" + `\.` + "`id`" + ` LIMIT \?`).
			WithArgs(prefix + "%", 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "quote_code"})) // 返回空行

		code, err := GetNewQuoteCode(context.Background(), gormDB)
		if err != nil {
			t.Errorf("期望没有错误，但是得到: %v", err)
		}
		expected := prefix + "-001"
		if code != expected {
			t.Errorf("期望 code 为 %s, 得到 %s", expected, code)
		}
	})

	t.Run("数据库有一条记录 BJ-YYYYMMDD-001，应该返回 BJ-YYYYMMDD-002", func(t *testing.T) {
		existingCode := prefix + "-001"
		rows := sqlmock.NewRows([]string{"id", "quote_code"}).
			AddRow(1, existingCode)

		mock.ExpectQuery(`SELECT \* FROM ` + "`quote`" + ` WHERE ` + "`quote`" + `\.` + "`quote_code`" + ` LIKE \? ORDER BY ` + "`quote`" + `\.` + "`quote_code`" + ` DESC,` + "`quote`" + `\.` + "`id`" + ` LIMIT \?`).
			WithArgs(prefix + "%", 1).
			WillReturnRows(rows)

		code, err := GetNewQuoteCode(context.Background(), gormDB)
		if err != nil {
			t.Errorf("期望没有错误，但是得到: %v", err)
		}
		expected := prefix + "-002"
		if code != expected {
			t.Errorf("期望 code 为 %s, 得到 %s", expected, code)
		}
	})

	t.Run("数据库有记录 BJ-YYYYMMDD-015，应该返回 BJ-YYYYMMDD-016", func(t *testing.T) {
		existingCode := prefix + "-015"
		rows := sqlmock.NewRows([]string{"id", "quote_code"}).
			AddRow(15, existingCode)

		mock.ExpectQuery(`SELECT \* FROM ` + "`quote`" + ` WHERE ` + "`quote`" + `\.` + "`quote_code`" + ` LIKE \? ORDER BY ` + "`quote`" + `\.` + "`quote_code`" + ` DESC,` + "`quote`" + `\.` + "`id`" + ` LIMIT \?`).
			WithArgs(prefix + "%", 1).
			WillReturnRows(rows)

		code, err := GetNewQuoteCode(context.Background(), gormDB)
		if err != nil {
			t.Errorf("期望没有错误，但是得到: %v", err)
		}
		expected := prefix + "-016"
		if code != expected {
			t.Errorf("期望 code 为 %s, 得到 %s", expected, code)
		}
	})
}
