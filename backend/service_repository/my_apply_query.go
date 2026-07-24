package service_repository

import (
	"context"

	"csun_server-backend/dao/model/quote_manage"
	quote_query "csun_server-backend/dao/query/quote_manage"
)

// MyApplyQueryResult 定义我的申请查询结果的数据结构
type MyApplyQueryResult struct {
	Quotes            []*quote_manage.Quote            `json:"quotes"`
	QuoteItems        []*quote_manage.AQuoteItem       `json:"quote_items"`
	QuoteProcesses    []*quote_manage.QuoteProcess     `json:"quote_processes"`
	QuoteProcessNodes []*quote_manage.QuoteProcessNode `json:"quote_process_nodes"`
}

// MyApplyQueryService 我的申请查询服务接口
type MyApplyQueryService interface {
	MyApplyQuery(ctx context.Context, userID int32) (*MyApplyQueryResult, error)
}

type myApplyQueryRepository interface {
	GetQuotesByCreatorID(ctx context.Context, creatorID int32) ([]*quote_manage.Quote, error)
	GetQuoteItemsByQuoteIDs(ctx context.Context, quoteIDs []int32) ([]*quote_manage.AQuoteItem, error)
	GetProcessesByQuoteIDs(ctx context.Context, quoteIDs []int32) ([]*quote_manage.QuoteProcess, error)
	GetNodesByProcessIDs(ctx context.Context, processIDs []int32) ([]*quote_manage.QuoteProcessNode, error)
}

type defaultMyApplyQueryRepository struct {
	db *DBEngine
}

func (r *defaultMyApplyQueryRepository) GetQuotesByCreatorID(ctx context.Context, creatorID int32) ([]*quote_manage.Quote, error) {
	q := quote_query.Use(r.db.QuoteManage)
	return q.Quote.WithContext(ctx).Where(q.Quote.CreatorID.Eq(creatorID)).Order(q.Quote.ID.Desc()).Find()
}

func (r *defaultMyApplyQueryRepository) GetQuoteItemsByQuoteIDs(ctx context.Context, quoteIDs []int32) ([]*quote_manage.AQuoteItem, error) {
	if len(quoteIDs) == 0 {
		return []*quote_manage.AQuoteItem{}, nil
	}
	q := quote_query.Use(r.db.QuoteManage)
	return q.AQuoteItem.WithContext(ctx).Where(q.AQuoteItem.QuoteID.In(quoteIDs...)).Order(q.AQuoteItem.ID.Desc()).Find()
}

func (r *defaultMyApplyQueryRepository) GetProcessesByQuoteIDs(ctx context.Context, quoteIDs []int32) ([]*quote_manage.QuoteProcess, error) {
	if len(quoteIDs) == 0 {
		return []*quote_manage.QuoteProcess{}, nil
	}
	q := quote_query.Use(r.db.QuoteManage)
	return q.QuoteProcess.WithContext(ctx).Where(q.QuoteProcess.QuoteID.In(quoteIDs...)).Order(q.QuoteProcess.ID.Desc()).Find()
}

func (r *defaultMyApplyQueryRepository) GetNodesByProcessIDs(ctx context.Context, processIDs []int32) ([]*quote_manage.QuoteProcessNode, error) {
	if len(processIDs) == 0 {
		return []*quote_manage.QuoteProcessNode{}, nil
	}
	q := quote_query.Use(r.db.QuoteManage)
	return q.QuoteProcessNode.WithContext(ctx).Where(q.QuoteProcessNode.ProcessID.In(processIDs...)).Order(q.QuoteProcessNode.ID.Desc()).Find()
}

type myApplyQueryService struct {
	repo myApplyQueryRepository
}

func NewMyApplyQueryService(db *DBEngine) MyApplyQueryService {
	return &myApplyQueryService{
		repo: &defaultMyApplyQueryRepository{db: db},
	}
}

func NewMockMyApplyQueryService(repo myApplyQueryRepository) MyApplyQueryService {
	return &myApplyQueryService{
		repo: repo,
	}
}

func (s *myApplyQueryService) MyApplyQuery(ctx context.Context, userID int32) (*MyApplyQueryResult, error) {
	result := &MyApplyQueryResult{
		Quotes:            make([]*quote_manage.Quote, 0),
		QuoteItems:        make([]*quote_manage.AQuoteItem, 0),
		QuoteProcesses:    make([]*quote_manage.QuoteProcess, 0),
		QuoteProcessNodes: make([]*quote_manage.QuoteProcessNode, 0),
	}

	// 1. 查quote：user.id == quote.creator_id的所有quote记录；拿到这些quote记录的“quote.id列表”
	quotes, err := s.repo.GetQuotesByCreatorID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(quotes) == 0 {
		return result, nil
	}
	result.Quotes = quotes

	quoteIDs := make([]int32, len(quotes))
	for i, q := range quotes {
		quoteIDs[i] = q.ID
	}

	// 2. 查A_quote_item：A_quote_item.quote_id in quote.id列表
	items, err := s.repo.GetQuoteItemsByQuoteIDs(ctx, quoteIDs)
	if err != nil {
		return nil, err
	}
	if len(items) > 0 {
		result.QuoteItems = items
	}

	// 3. 查quote_process：quote_process.quote_id in quote.id列表；拿到这些quote_process记录的“quote_process.id列表”
	processes, err := s.repo.GetProcessesByQuoteIDs(ctx, quoteIDs)
	if err != nil {
		return nil, err
	}
	if len(processes) > 0 {
		result.QuoteProcesses = processes
	}

	processIDs := make([]int32, 0, len(processes))
	for _, p := range processes {
		processIDs = append(processIDs, p.ID)
	}

	// 4. 查quote_process_node：quote_process_node.process_id in quote_process.id列表
	if len(processIDs) > 0 {
		nodes, err := s.repo.GetNodesByProcessIDs(ctx, processIDs)
		if err != nil {
			return nil, err
		}
		if len(nodes) > 0 {
			result.QuoteProcessNodes = nodes
		}
	}

	return result, nil
}
