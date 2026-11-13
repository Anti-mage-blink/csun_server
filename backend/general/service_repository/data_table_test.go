package service_repository

import (
	"context"
	"errors"
	"testing"
)

type mockDataTableService struct {
	QueryListFn    func(ctx context.Context, tableStr string, relationTables []string) (*QueryResult, error)
	CreateRecordFn func(ctx context.Context, tableStr string, record map[string]interface{}) (map[string]interface{}, error)
	UpdateRecordFn func(ctx context.Context, tableStr string, id interface{}, record map[string]interface{}) error
	DeleteRecordFn func(ctx context.Context, tableStr string, id interface{}) error
}

func (m *mockDataTableService) ParseDBAndTable(fullStr string) (string, string, error) {
	return ParseDBAndTable(fullStr)
}

func (m *mockDataTableService) QueryList(ctx context.Context, tableStr string, relationTables []string) (*QueryResult, error) {
	if m.QueryListFn != nil {
		return m.QueryListFn(ctx, tableStr, relationTables)
	}
	return &QueryResult{}, nil
}

func (m *mockDataTableService) CreateRecord(ctx context.Context, tableStr string, record map[string]interface{}) (map[string]interface{}, error) {
	if m.CreateRecordFn != nil {
		return m.CreateRecordFn(ctx, tableStr, record)
	}
	return record, nil
}

func (m *mockDataTableService) UpdateRecord(ctx context.Context, tableStr string, id interface{}, record map[string]interface{}) error {
	if m.UpdateRecordFn != nil {
		return m.UpdateRecordFn(ctx, tableStr, id, record)
	}
	return nil
}

func (m *mockDataTableService) DeleteRecord(ctx context.Context, tableStr string, id interface{}) error {
	if m.DeleteRecordFn != nil {
		return m.DeleteRecordFn(ctx, tableStr, id)
	}
	return nil
}

func TestParseDBAndTable(t *testing.T) {
	// 1. 正常带库名和表名
	db, tbl, err := ParseDBAndTable("general.customer")
	if err != nil || db != "general" || tbl != "customer" {
		t.Fatalf("expected general.customer, got db=%s, tbl=%s, err=%v", db, tbl, err)
	}

	// 2. 仅有表名
	db, tbl, err = ParseDBAndTable("quote_process")
	if err != nil || db != "" || tbl != "quote_process" {
		t.Fatalf("expected quote_process, got db=%s, tbl=%s, err=%v", db, tbl, err)
	}

	// 3. 错误格式
	_, _, err = ParseDBAndTable("")
	if err == nil {
		t.Fatal("expected error for empty string, got nil")
	}

	_, _, err = ParseDBAndTable("general.customer.extra")
	if err == nil {
		t.Fatal("expected error for extra dots, got nil")
	}

	_, _, err = ParseDBAndTable("general; DROP TABLE test;--")
	if err == nil {
		t.Fatal("expected error for SQL injection string, got nil")
	}
}

func TestDataTableService_MockCRUD(t *testing.T) {
	ctx := context.Background()
	queryCalled := false
	createCalled := false
	updateCalled := false
	deleteCalled := false

	mockSvc := &mockDataTableService{
		QueryListFn: func(ctx context.Context, tableStr string, relationTables []string) (*QueryResult, error) {
			queryCalled = true
			if tableStr != "general.customer" {
				return nil, errors.New("table mismatch")
			}
			return &QueryResult{
				Main: []map[string]interface{}{{"id": 1, "company_name": "测试客户"}},
				Relations: map[string][]map[string]interface{}{
					"quote_manage.product_spec": {{"id": 10, "product_name": "产品A"}},
				},
			}, nil
		},
		CreateRecordFn: func(ctx context.Context, tableStr string, record map[string]interface{}) (map[string]interface{}, error) {
			createCalled = true
			record["id"] = 100
			return record, nil
		},
		UpdateRecordFn: func(ctx context.Context, tableStr string, id interface{}, record map[string]interface{}) error {
			updateCalled = true
			if id == nil {
				return errors.New("nil id")
			}
			return nil
		},
		DeleteRecordFn: func(ctx context.Context, tableStr string, id interface{}) error {
			deleteCalled = true
			return nil
		},
	}

	// 1. 测试查 (主表 + 关联表)
	res, err := mockSvc.QueryList(ctx, "general.customer", []string{"quote_manage.product_spec"})
	if err != nil || len(res.Main) != 1 || len(res.Relations["quote_manage.product_spec"]) != 1 || !queryCalled {
		t.Fatalf("QueryList failed, err: %v", err)
	}

	// 2. 测试增
	rec, err := mockSvc.CreateRecord(ctx, "general.customer", map[string]interface{}{"company_name": "新客户"})
	if err != nil || rec["id"] != 100 || !createCalled {
		t.Fatalf("CreateRecord failed, err: %v", err)
	}

	// 3. 测试改
	err = mockSvc.UpdateRecord(ctx, "general.customer", 100, map[string]interface{}{"company_name": "改名客户"})
	if err != nil || !updateCalled {
		t.Fatalf("UpdateRecord failed, err: %v", err)
	}

	// 4. 测试删
	err = mockSvc.DeleteRecord(ctx, "general.customer", 100)
	if err != nil || !deleteCalled {
		t.Fatalf("DeleteRecord failed, err: %v", err)
	}
}
