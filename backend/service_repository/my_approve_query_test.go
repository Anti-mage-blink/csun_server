package service_repository

import (
"context"
"errors"
"testing"

"csun_server-backend/dao/model/quote_manage"
)

// MockMyApproveQueryRepository 我的审批查询的 Mock 仓储实现
type MockMyApproveQueryRepository struct {
GetProcessesByApproverIDFn func(ctx context.Context, approverID int32) ([]*quote_manage.QuoteProcess, error)
GetNodesByProcessIDsFn     func(ctx context.Context, processIDs []int32) ([]*quote_manage.QuoteProcessNode, error)
GetQuotesByQuoteIDsFn      func(ctx context.Context, quoteIDs []int32) ([]*quote_manage.Quote, error)
GetQuoteItemsByQuoteIDsFn  func(ctx context.Context, quoteIDs []int32) ([]*quote_manage.AQuoteItem, error)
}

func (m *MockMyApproveQueryRepository) GetProcessesByApproverID(ctx context.Context, approverID int32) ([]*quote_manage.QuoteProcess, error) {
return m.GetProcessesByApproverIDFn(ctx, approverID)
}

func (m *MockMyApproveQueryRepository) GetNodesByProcessIDs(ctx context.Context, processIDs []int32) ([]*quote_manage.QuoteProcessNode, error) {
return m.GetNodesByProcessIDsFn(ctx, processIDs)
}

func (m *MockMyApproveQueryRepository) GetQuotesByQuoteIDs(ctx context.Context, quoteIDs []int32) ([]*quote_manage.Quote, error) {
return m.GetQuotesByQuoteIDsFn(ctx, quoteIDs)
}

func (m *MockMyApproveQueryRepository) GetQuoteItemsByQuoteIDs(ctx context.Context, quoteIDs []int32) ([]*quote_manage.AQuoteItem, error) {
return m.GetQuoteItemsByQuoteIDsFn(ctx, quoteIDs)
}

// TestMyApproveQuery_NoProcesses 测试没有任何审批流程记录的情况
func TestMyApproveQuery_NoProcesses(t *testing.T) {
mock := &MockMyApproveQueryRepository{
GetProcessesByApproverIDFn: func(ctx context.Context, approverID int32) ([]*quote_manage.QuoteProcess, error) {
return nil, nil
},
}
svc := NewMockMyApproveQueryService(mock)
res, err := svc.MyApproveQuery(context.Background(), 1)
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

// TestMyApproveQuery_DBError 测试底层数据库查询出错的情况
func TestMyApproveQuery_DBError(t *testing.T) {
mock := &MockMyApproveQueryRepository{
GetProcessesByApproverIDFn: func(ctx context.Context, approverID int32) ([]*quote_manage.QuoteProcess, error) {
return nil, errors.New("db query error")
},
}
svc := NewMockMyApproveQueryService(mock)
_, err := svc.MyApproveQuery(context.Background(), 1)
if err == nil {
t.Fatal("expected error, got nil")
}
if err.Error() != "db query error" {
t.Errorf("expected 'db query error', got '%s'", err.Error())
}
}

// TestMyApproveQuery_Success 测试成功查到审批流程，并组装关联数据
func TestMyApproveQuery_Success(t *testing.T) {
processID := int32(100)
quoteID := int32(200)
approverID := int32(1)

process := &quote_manage.QuoteProcess{
ID:         processID,
QuoteID:    &quoteID,
ApproverID: &approverID,
}

mock := &MockMyApproveQueryRepository{
GetProcessesByApproverIDFn: func(ctx context.Context, aid int32) ([]*quote_manage.QuoteProcess, error) {
if aid == approverID {
return []*quote_manage.QuoteProcess{process}, nil
}
return nil, nil
},
GetNodesByProcessIDsFn: func(ctx context.Context, pids []int32) ([]*quote_manage.QuoteProcessNode, error) {
if len(pids) == 1 && pids[0] == processID {
return []*quote_manage.QuoteProcessNode{
{ID: 1, ProcessID: &processID},
{ID: 2, ProcessID: &processID},
}, nil
}
return nil, nil
},
GetQuotesByQuoteIDsFn: func(ctx context.Context, qids []int32) ([]*quote_manage.Quote, error) {
if len(qids) == 1 && qids[0] == quoteID {
return []*quote_manage.Quote{
{ID: quoteID},
}, nil
}
return nil, nil
},
GetQuoteItemsByQuoteIDsFn: func(ctx context.Context, qids []int32) ([]*quote_manage.AQuoteItem, error) {
if len(qids) == 1 && qids[0] == quoteID {
return []*quote_manage.AQuoteItem{
{ID: 1, QuoteID: &quoteID},
{ID: 2, QuoteID: &quoteID},
}, nil
}
return nil, nil
},
}

svc := NewMockMyApproveQueryService(mock)
res, err := svc.MyApproveQuery(context.Background(), approverID)
if err != nil {
t.Fatalf("expected no error, got %v", err)
}

if res.Total != 1 {
t.Errorf("expected total 1, got %d", res.Total)
}
if len(res.QuoteProcesses) != 1 {
t.Errorf("expected 1 process, got %d", len(res.QuoteProcesses))
}
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
