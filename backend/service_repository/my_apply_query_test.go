package service_repository

import (
	"context"
	"errors"
	"testing"

	"csun_server-backend/dao/model/quote_manage"
)

// MockMyApplyQueryRepository 我的申请查询仓储的 Mock 实现
type MockMyApplyQueryRepository struct {
	GetQuotesByCreatorIDFn    func(ctx context.Context, creatorID int32) ([]*quote_manage.Quote, error)
	GetQuoteItemsByQuoteIDsFn func(ctx context.Context, quoteIDs []int32) ([]*quote_manage.AQuoteItem, error)
	GetProcessesByQuoteIDsFn  func(ctx context.Context, quoteIDs []int32) ([]*quote_manage.QuoteProcess, error)
	GetNodesByProcessIDsFn    func(ctx context.Context, processIDs []int32) ([]*quote_manage.QuoteProcessNode, error)
}

func (m *MockMyApplyQueryRepository) GetQuotesByCreatorID(ctx context.Context, creatorID int32) ([]*quote_manage.Quote, error) {
	return m.GetQuotesByCreatorIDFn(ctx, creatorID)
}

func (m *MockMyApplyQueryRepository) GetQuoteItemsByQuoteIDs(ctx context.Context, quoteIDs []int32) ([]*quote_manage.AQuoteItem, error) {
	return m.GetQuoteItemsByQuoteIDsFn(ctx, quoteIDs)
}

func (m *MockMyApplyQueryRepository) GetProcessesByQuoteIDs(ctx context.Context, quoteIDs []int32) ([]*quote_manage.QuoteProcess, error) {
	return m.GetProcessesByQuoteIDsFn(ctx, quoteIDs)
}

func (m *MockMyApplyQueryRepository) GetNodesByProcessIDs(ctx context.Context, processIDs []int32) ([]*quote_manage.QuoteProcessNode, error) {
	return m.GetNodesByProcessIDsFn(ctx, processIDs)
}

// TestMyApplyQuery_Success 测试成功获取我的申请数据
func TestMyApplyQuery_Success(t *testing.T) {
	mockQuotes := []*quote_manage.Quote{{ID: 10}}
	mockItems := []*quote_manage.AQuoteItem{{ID: 20, QuoteID: func() *int32 { id := int32(10); return &id }()}}
	mockProcesses := []*quote_manage.QuoteProcess{{ID: 30, QuoteID: func() *int32 { id := int32(10); return &id }()}}
	mockNodes := []*quote_manage.QuoteProcessNode{{ID: 40, ProcessID: func() *int32 { id := int32(30); return &id }()}}

	mock := &MockMyApplyQueryRepository{
		GetQuotesByCreatorIDFn: func(ctx context.Context, creatorID int32) ([]*quote_manage.Quote, error) {
			if creatorID != 1 {
				t.Errorf("expected creatorID 1, got %d", creatorID)
			}
			return mockQuotes, nil
		},
		GetQuoteItemsByQuoteIDsFn: func(ctx context.Context, quoteIDs []int32) ([]*quote_manage.AQuoteItem, error) {
			if len(quoteIDs) != 1 || quoteIDs[0] != 10 {
				t.Errorf("expected quoteIDs [10], got %v", quoteIDs)
			}
			return mockItems, nil
		},
		GetProcessesByQuoteIDsFn: func(ctx context.Context, quoteIDs []int32) ([]*quote_manage.QuoteProcess, error) {
			if len(quoteIDs) != 1 || quoteIDs[0] != 10 {
				t.Errorf("expected quoteIDs [10], got %v", quoteIDs)
			}
			return mockProcesses, nil
		},
		GetNodesByProcessIDsFn: func(ctx context.Context, processIDs []int32) ([]*quote_manage.QuoteProcessNode, error) {
			if len(processIDs) != 1 || processIDs[0] != 30 {
				t.Errorf("expected processIDs [30], got %v", processIDs)
			}
			return mockNodes, nil
		},
	}

	svc := NewMockMyApplyQueryService(mock)
	res, err := svc.MyApplyQuery(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(res.Quotes) != 1 || res.Quotes[0].ID != 10 {
		t.Errorf("quotes slice unexpected")
	}
	if len(res.QuoteItems) != 1 || res.QuoteItems[0].ID != 20 {
		t.Errorf("quote items slice unexpected")
	}
	if len(res.QuoteProcesses) != 1 || res.QuoteProcesses[0].ID != 30 {
		t.Errorf("processes slice unexpected")
	}
	if len(res.QuoteProcessNodes) != 1 || res.QuoteProcessNodes[0].ID != 40 {
		t.Errorf("nodes slice unexpected")
	}
}

// TestMyApplyQuery_NoQuotes 测试无报价单（我的申请为空）的情况
func TestMyApplyQuery_NoQuotes(t *testing.T) {
	mock := &MockMyApplyQueryRepository{
		GetQuotesByCreatorIDFn: func(ctx context.Context, creatorID int32) ([]*quote_manage.Quote, error) {
			return []*quote_manage.Quote{}, nil
		},
	}

	svc := NewMockMyApplyQueryService(mock)
	res, err := svc.MyApplyQuery(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(res.Quotes) != 0 || len(res.QuoteItems) != 0 || len(res.QuoteProcesses) != 0 || len(res.QuoteProcessNodes) != 0 {
		t.Errorf("expected empty slices for all fields, got %+v", res)
	}
}

// TestMyApplyQuery_DBError 测试数据库出错的情况
func TestMyApplyQuery_DBError(t *testing.T) {
	mock := &MockMyApplyQueryRepository{
		GetQuotesByCreatorIDFn: func(ctx context.Context, creatorID int32) ([]*quote_manage.Quote, error) {
			return nil, errors.New("db error")
		},
	}

	svc := NewMockMyApplyQueryService(mock)
	_, err := svc.MyApplyQuery(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "db error" {
		t.Errorf("expected 'db error', got '%s'", err.Error())
	}
}
