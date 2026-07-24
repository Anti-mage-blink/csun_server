package service_repository

import (
	"context"
	"errors"
	"testing"

	"csun_server-backend/dao/model/quote_manage"
)

// MockFilingLookRepository 备案查看仓储的 Mock 实现
type MockFilingLookRepository struct {
	GetAllProcessesFn  func(ctx context.Context) ([]*quote_manage.QuoteProcess, error)
	GetAllNodesFn      func(ctx context.Context) ([]*quote_manage.QuoteProcessNode, error)
	GetAllQuotesFn     func(ctx context.Context) ([]*quote_manage.Quote, error)
	GetAllQuoteItemsFn func(ctx context.Context) ([]*quote_manage.AQuoteItem, error)
}

func (m *MockFilingLookRepository) GetAllProcesses(ctx context.Context) ([]*quote_manage.QuoteProcess, error) {
	return m.GetAllProcessesFn(ctx)
}

func (m *MockFilingLookRepository) GetAllNodes(ctx context.Context) ([]*quote_manage.QuoteProcessNode, error) {
	return m.GetAllNodesFn(ctx)
}

func (m *MockFilingLookRepository) GetAllQuotes(ctx context.Context) ([]*quote_manage.Quote, error) {
	return m.GetAllQuotesFn(ctx)
}

func (m *MockFilingLookRepository) GetAllQuoteItems(ctx context.Context) ([]*quote_manage.AQuoteItem, error) {
	return m.GetAllQuoteItemsFn(ctx)
}

// TestFilingLook_Success 测试成功获取全量备案数据的情况
func TestFilingLook_Success(t *testing.T) {
	mockProcesses := []*quote_manage.QuoteProcess{{ID: 1}}
	mockNodes := []*quote_manage.QuoteProcessNode{{ID: 1}}
	mockQuotes := []*quote_manage.Quote{{ID: 1}}
	mockItems := []*quote_manage.AQuoteItem{{ID: 1}}

	mock := &MockFilingLookRepository{
		GetAllProcessesFn: func(ctx context.Context) ([]*quote_manage.QuoteProcess, error) {
			return mockProcesses, nil
		},
		GetAllNodesFn: func(ctx context.Context) ([]*quote_manage.QuoteProcessNode, error) {
			return mockNodes, nil
		},
		GetAllQuotesFn: func(ctx context.Context) ([]*quote_manage.Quote, error) {
			return mockQuotes, nil
		},
		GetAllQuoteItemsFn: func(ctx context.Context) ([]*quote_manage.AQuoteItem, error) {
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
	if len(res.Quotes) != 1 || res.Quotes[0].ID != 1 {
		t.Errorf("quotes slice unexpected")
	}
	if len(res.QuoteItems) != 1 || res.QuoteItems[0].ID != 1 {
		t.Errorf("quote items slice unexpected")
	}
}

// TestFilingLook_DBError 测试其中一个查询数据库出错的情况
func TestFilingLook_DBError(t *testing.T) {
	mock := &MockFilingLookRepository{
		GetAllProcessesFn: func(ctx context.Context) ([]*quote_manage.QuoteProcess, error) {
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
