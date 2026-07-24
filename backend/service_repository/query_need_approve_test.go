package service_repository

import (
	"context"
	"errors"
	"testing"

	"csun_server-backend/dao/model/quote_manage"
)

// MockQueryNeedApproveRepository 待审批查询的 Mock 仓储实现
type MockQueryNeedApproveRepository struct {
	GetPendingNodesByEmployeeIDFn func(ctx context.Context, employeeID int32) ([]*quote_manage.QuoteProcessNode, error)
	GetProcessByIDFn              func(ctx context.Context, processID int32) (*quote_manage.QuoteProcess, error)
	GetNodesByProcessIDFn         func(ctx context.Context, processID int32) ([]*quote_manage.QuoteProcessNode, error)
	GetQuoteByIDFn                func(ctx context.Context, quoteID int32) (*quote_manage.Quote, error)
	GetQuoteItemsByQuoteIDFn      func(ctx context.Context, quoteID int32) ([]*quote_manage.AQuoteItem, error)
}

func (m *MockQueryNeedApproveRepository) GetPendingNodesByEmployeeID(ctx context.Context, employeeID int32) ([]*quote_manage.QuoteProcessNode, error) {
	return m.GetPendingNodesByEmployeeIDFn(ctx, employeeID)
}

func (m *MockQueryNeedApproveRepository) GetProcessByID(ctx context.Context, processID int32) (*quote_manage.QuoteProcess, error) {
	return m.GetProcessByIDFn(ctx, processID)
}

func (m *MockQueryNeedApproveRepository) GetNodesByProcessID(ctx context.Context, processID int32) ([]*quote_manage.QuoteProcessNode, error) {
	return m.GetNodesByProcessIDFn(ctx, processID)
}

func (m *MockQueryNeedApproveRepository) GetQuoteByID(ctx context.Context, quoteID int32) (*quote_manage.Quote, error) {
	return m.GetQuoteByIDFn(ctx, quoteID)
}

func (m *MockQueryNeedApproveRepository) GetQuoteItemsByQuoteID(ctx context.Context, quoteID int32) ([]*quote_manage.AQuoteItem, error) {
	return m.GetQuoteItemsByQuoteIDFn(ctx, quoteID)
}

// TestQueryNeedApprove_NoPending 测试没有任何待审批审批节点的情况
func TestQueryNeedApprove_NoPending(t *testing.T) {
	mock := &MockQueryNeedApproveRepository{
		GetPendingNodesByEmployeeIDFn: func(ctx context.Context, employeeID int32) ([]*quote_manage.QuoteProcessNode, error) {
			return nil, nil
		},
	}
	svc := NewMockQueryNeedApproveService(mock)
	res, err := svc.QueryNeedApprove(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Total != 0 {
		t.Errorf("expected total 0, got %d", res.Total)
	}
	if len(res.QuoteProcesses) != 0 || len(res.QuoteProcessNodes) != 0 || len(res.Quotes) != 0 || len(res.QuoteItems) != 0 {
		t.Errorf("expected all lists to be empty, got: processes=%d, nodes=%d, quotes=%d, items=%d",
			len(res.QuoteProcesses), len(res.QuoteProcessNodes), len(res.Quotes), len(res.QuoteItems))
	}
}

// TestQueryNeedApprove_DBError 测试底层数据库查询出错的情况
func TestQueryNeedApprove_DBError(t *testing.T) {
	mock := &MockQueryNeedApproveRepository{
		GetPendingNodesByEmployeeIDFn: func(ctx context.Context, employeeID int32) ([]*quote_manage.QuoteProcessNode, error) {
			return nil, errors.New("db query error")
		},
	}
	svc := NewMockQueryNeedApproveService(mock)
	_, err := svc.QueryNeedApprove(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "db query error" {
		t.Errorf("expected 'db query error', got '%s'", err.Error())
	}
}

// TestQueryNeedApprove_Success 测试成功查到待审批，并组装关联数据、验证去重逻辑
func TestQueryNeedApprove_Success(t *testing.T) {
	processID := int32(100)
	quoteID := int32(200)

	node1 := &quote_manage.QuoteProcessNode{
		ID:        1,
		ProcessID: &processID,
	}

	mock := &MockQueryNeedApproveRepository{
		GetPendingNodesByEmployeeIDFn: func(ctx context.Context, employeeID int32) ([]*quote_manage.QuoteProcessNode, error) {
			return []*quote_manage.QuoteProcessNode{node1}, nil
		},
		GetProcessByIDFn: func(ctx context.Context, pid int32) (*quote_manage.QuoteProcess, error) {
			if pid == processID {
				return &quote_manage.QuoteProcess{
					ID:      processID,
					QuoteID: &quoteID,
				}, nil
			}
			return nil, nil
		},
		GetNodesByProcessIDFn: func(ctx context.Context, pid int32) ([]*quote_manage.QuoteProcessNode, error) {
			if pid == processID {
				// 返回两个 node，包含 node1 本身，校验去重逻辑
				return []*quote_manage.QuoteProcessNode{
					node1,
					{
						ID:        2,
						ProcessID: &processID,
					},
				}, nil
			}
			return nil, nil
		},
		GetQuoteByIDFn: func(ctx context.Context, qid int32) (*quote_manage.Quote, error) {
			if qid == quoteID {
				return &quote_manage.Quote{
					ID: quoteID,
				}, nil
			}
			return nil, nil
		},
		GetQuoteItemsByQuoteIDFn: func(ctx context.Context, qid int32) ([]*quote_manage.AQuoteItem, error) {
			if qid == quoteID {
				return []*quote_manage.AQuoteItem{
					{
						ID:      1,
						QuoteID: &quoteID,
					},
					{
						ID:      2,
						QuoteID: &quoteID,
					},
				}, nil
			}
			return nil, nil
		},
	}

	svc := NewMockQueryNeedApproveService(mock)
	res, err := svc.QueryNeedApprove(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if res.Total != 1 {
		t.Errorf("expected total 1, got %d", res.Total)
	}
	if len(res.QuoteProcesses) != 1 {
		t.Errorf("expected 1 process, got %d", len(res.QuoteProcesses))
	}
	// 期望 2 个节点
	if len(res.QuoteProcessNodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(res.QuoteProcessNodes))
	}
	if len(res.Quotes) != 1 {
		t.Errorf("expected 1 quote, got %d", len(res.Quotes))
	}
	if len(res.QuoteItems) != 2 {
		t.Errorf("expected 2 quote items, got %d", len(res.QuoteItems))
	}
}
