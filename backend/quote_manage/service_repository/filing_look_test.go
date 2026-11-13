package service_repository

import (
	"context"
	"errors"
	"testing"

	"csun_server-backend/dao/model/quote_manage"
)

// MockFilingLookRepository 备案查看仓储的 Mock 实现
type MockFilingLookRepository struct {
	GetValidProcessesFn       func(ctx context.Context) ([]*quote_manage.QuoteProcess, error)
	GetNodesByProcessIDsFn    func(ctx context.Context, processIDs []int32) ([]*quote_manage.QuoteProcessNode, error)
	GetQuotesByQuoteIDsFn     func(ctx context.Context, quoteIDs []int32) ([]*quote_manage.Quote, error)
	GetQuoteItemsByQuoteIDsFn func(ctx context.Context, quoteIDs []int32) ([]*quote_manage.AQuoteItem, error)
}

func (m *MockFilingLookRepository) GetValidProcesses(ctx context.Context) ([]*quote_manage.QuoteProcess, error) {
	return m.GetValidProcessesFn(ctx)
}

func (m *MockFilingLookRepository) GetNodesByProcessIDs(ctx context.Context, processIDs []int32) ([]*quote_manage.QuoteProcessNode, error) {
	return m.GetNodesByProcessIDsFn(ctx, processIDs)
}

func (m *MockFilingLookRepository) GetQuotesByQuoteIDs(ctx context.Context, quoteIDs []int32) ([]*quote_manage.Quote, error) {
	return m.GetQuotesByQuoteIDsFn(ctx, quoteIDs)
}

func (m *MockFilingLookRepository) GetQuoteItemsByQuoteIDs(ctx context.Context, quoteIDs []int32) ([]*quote_manage.AQuoteItem, error) {
	return m.GetQuoteItemsByQuoteIDsFn(ctx, quoteIDs)
}

// TestFilingLook_Success 测试成功获取备案数据的情况
func TestFilingLook_Success(t *testing.T) {
	qid := int32(10)
	pid := int32(1)
	mockProcesses := []*quote_manage.QuoteProcess{{ID: 1, QuoteID: &qid}}
	mockNodes := []*quote_manage.QuoteProcessNode{{ID: 1, ProcessID: &pid}}
	mockQuotes := []*quote_manage.Quote{{ID: 10}}
	mockItems := []*quote_manage.AQuoteItem{{ID: 100, QuoteID: &qid}}

	mock := &MockFilingLookRepository{
		GetValidProcessesFn: func(ctx context.Context) ([]*quote_manage.QuoteProcess, error) {
			return mockProcesses, nil
		},
		GetNodesByProcessIDsFn: func(ctx context.Context, processIDs []int32) ([]*quote_manage.QuoteProcessNode, error) {
			return mockNodes, nil
		},
		GetQuotesByQuoteIDsFn: func(ctx context.Context, quoteIDs []int32) ([]*quote_manage.Quote, error) {
			return mockQuotes, nil
		},
		GetQuoteItemsByQuoteIDsFn: func(ctx context.Context, quoteIDs []int32) ([]*quote_manage.AQuoteItem, error) {
			return mockItems, nil
		},
	}

	svc := NewMockFilingLookService(mock)
	res, err := svc.FilingLook(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(res.QuoteProcesses) != 1 || res.QuoteProcesses[0].ID != 1 {
		t.Errorf("processes slice unexpected")
	}
	if len(res.QuoteProcessNodes) != 1 || res.QuoteProcessNodes[0].ID != 1 {
		t.Errorf("nodes slice unexpected")
	}
	if len(res.Quotes) != 1 || res.Quotes[0].ID != 10 {
		t.Errorf("quotes slice unexpected")
	}
	if len(res.QuoteItems) != 1 || res.QuoteItems[0].ID != 100 {
		t.Errorf("quote items slice unexpected")
	}
}

// TestFilingLook_DBError 测试其中一个查询数据库出错的情况
func TestFilingLook_DBError(t *testing.T) {
	mock := &MockFilingLookRepository{
		GetValidProcessesFn: func(ctx context.Context) ([]*quote_manage.QuoteProcess, error) {
			return nil, errors.New("db connection failure")
		},
	}

	svc := NewMockFilingLookService(mock)
	_, err := svc.FilingLook(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "db connection failure" {
		t.Errorf("expected 'db connection failure', got '%s'", err.Error())
	}
}
