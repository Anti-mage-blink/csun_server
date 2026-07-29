package service_repository

import (
	"context"

	"csun_server-backend/dao/model/quote_manage"
	quote_query "csun_server-backend/dao/query/quote_manage"
)

// MyApproveQueryResult 定义我的审批查询结果的数据结构
type MyApproveQueryResult struct {
	Total             int64                            `json:"total"`
	QuoteProcesses    []*quote_manage.QuoteProcess     `json:"quote_processes"`
	QuoteProcessNodes []*quote_manage.QuoteProcessNode `json:"quote_process_nodes"`
	Quotes            []*quote_manage.Quote            `json:"quotes"`
	QuoteItems        []*quote_manage.AQuoteItem       `json:"quote_items"`
}

// MyApproveQueryService 我的审批查询服务接口
type MyApproveQueryService interface {
	MyApproveQuery(ctx context.Context, userID int32) (*MyApproveQueryResult, error)
}

type myApproveQueryRepository interface {
	GetProcessesByApproverID(ctx context.Context, approverID int32) ([]*quote_manage.QuoteProcess, error)
	GetNodesByProcessIDs(ctx context.Context, processIDs []int32) ([]*quote_manage.QuoteProcessNode, error)
	GetQuotesByQuoteIDs(ctx context.Context, quoteIDs []int32) ([]*quote_manage.Quote, error)
	GetQuoteItemsByQuoteIDs(ctx context.Context, quoteIDs []int32) ([]*quote_manage.AQuoteItem, error)
}

type defaultMyApproveQueryRepository struct {
	db *DBEngine
}

func (r *defaultMyApproveQueryRepository) GetProcessesByApproverID(ctx context.Context, approverID int32) ([]*quote_manage.QuoteProcess, error) {
	q := quote_query.Use(r.db.QuoteManage)
	return q.QuoteProcess.WithContext(ctx).Where(
		q.QuoteProcess.ApproverID.Eq(approverID),
	).Order(q.QuoteProcess.ID.Desc()).Find()
}

func (r *defaultMyApproveQueryRepository) GetNodesByProcessIDs(ctx context.Context, processIDs []int32) ([]*quote_manage.QuoteProcessNode, error) {
	if len(processIDs) == 0 {
		return []*quote_manage.QuoteProcessNode{}, nil
	}
	q := quote_query.Use(r.db.QuoteManage)
	return q.QuoteProcessNode.WithContext(ctx).Where(
		q.QuoteProcessNode.ProcessID.In(processIDs...),
	).Order(q.QuoteProcessNode.ID.Desc()).Find()
}

func (r *defaultMyApproveQueryRepository) GetQuotesByQuoteIDs(ctx context.Context, quoteIDs []int32) ([]*quote_manage.Quote, error) {
	if len(quoteIDs) == 0 {
		return []*quote_manage.Quote{}, nil
	}
	q := quote_query.Use(r.db.QuoteManage)
	return q.Quote.WithContext(ctx).Where(
		q.Quote.ID.In(quoteIDs...),
	).Order(q.Quote.ID.Desc()).Find()
}

func (r *defaultMyApproveQueryRepository) GetQuoteItemsByQuoteIDs(ctx context.Context, quoteIDs []int32) ([]*quote_manage.AQuoteItem, error) {
	if len(quoteIDs) == 0 {
		return []*quote_manage.AQuoteItem{}, nil
	}
	q := quote_query.Use(r.db.QuoteManage)
	return q.AQuoteItem.WithContext(ctx).Where(
		q.AQuoteItem.QuoteID.In(quoteIDs...),
	).Order(q.AQuoteItem.ID.Desc()).Find()
}

type myApproveQueryService struct {
	repo myApproveQueryRepository
}

func NewMyApproveQueryService(db *DBEngine) MyApproveQueryService {
	return &myApproveQueryService{
		repo: &defaultMyApproveQueryRepository{db: db},
	}
}

func NewMockMyApproveQueryService(repo myApproveQueryRepository) MyApproveQueryService {
	return &myApproveQueryService{
		repo: repo,
	}
}

func (s *myApproveQueryService) MyApproveQuery(ctx context.Context, userID int32) (*MyApproveQueryResult, error) {
	// 初始化返回结构体，默认给空切片，避免前端接收到 null
	result := &MyApproveQueryResult{
		Total:             0,
		QuoteProcesses:    make([]*quote_manage.QuoteProcess, 0),
		QuoteProcessNodes: make([]*quote_manage.QuoteProcessNode, 0),
		Quotes:            make([]*quote_manage.Quote, 0),
		QuoteItems:        make([]*quote_manage.AQuoteItem, 0),
	}

	// 1. 查quote_process：user.id == approver_id
	processes, err := s.repo.GetProcessesByApproverID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(processes) == 0 {
		return result, nil
	}

	result.QuoteProcesses = processes
	result.Total = int64(len(processes))

	// 拿到这些quote_process记录的“quote_process.id列表”和“quote_id列表”
	processIDs := make([]int32, 0, len(processes))
	quoteIDMap := make(map[int32]bool)
	quoteIDs := make([]int32, 0, len(processes))

	for _, p := range processes {
		if p == nil {
			continue
		}
		processIDs = append(processIDs, p.ID)
		if p.QuoteID != nil && !quoteIDMap[*p.QuoteID] {
			quoteIDMap[*p.QuoteID] = true
			quoteIDs = append(quoteIDs, *p.QuoteID)
		}
	}

	// 2. 查quote_process_node：quote_process_node.process_id in quote_process.id列表
	if len(processIDs) > 0 {
		nodes, err := s.repo.GetNodesByProcessIDs(ctx, processIDs)
		if err != nil {
			return nil, err
		}
		if len(nodes) > 0 {
			result.QuoteProcessNodes = nodes
		}
	}

	// 3. 查quote：quote.id in quote.id列表
	if len(quoteIDs) > 0 {
		quotes, err := s.repo.GetQuotesByQuoteIDs(ctx, quoteIDs)
		if err != nil {
			return nil, err
		}
		if len(quotes) > 0 {
			result.Quotes = quotes
		}

		// 4. 查A_quote_item：A_quote_item.quote_id in quote.id列表
		items, err := s.repo.GetQuoteItemsByQuoteIDs(ctx, quoteIDs)
		if err != nil {
			return nil, err
		}
		if len(items) > 0 {
			result.QuoteItems = items
		}
	}

	return result, nil
}
