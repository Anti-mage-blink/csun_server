package service_repository

import (
	"context"

	"csun_server-backend/dao/model/quote_manage"
	quote_query "csun_server-backend/dao/query/quote_manage"
)

// FilingLookResult 定义备案查看结果的数据结构
type FilingLookResult struct {
	QuoteProcesses    []*quote_manage.QuoteProcess     `json:"quote_processes"`
	QuoteProcessNodes []*quote_manage.QuoteProcessNode `json:"quote_process_nodes"`
	Quotes            []*quote_manage.Quote            `json:"quotes"`
	QuoteItems        []*quote_manage.AQuoteItem       `json:"quote_items"`
}

// FilingLookService 备案查看服务接口
type FilingLookService interface {
	FilingLook(ctx context.Context) (*FilingLookResult, error)
}

type filingLookRepository interface {
	GetValidProcesses(ctx context.Context) ([]*quote_manage.QuoteProcess, error)
	GetNodesByProcessIDs(ctx context.Context, processIDs []int32) ([]*quote_manage.QuoteProcessNode, error)
	GetQuotesByQuoteIDs(ctx context.Context, quoteIDs []int32) ([]*quote_manage.Quote, error)
	GetQuoteItemsByQuoteIDs(ctx context.Context, quoteIDs []int32) ([]*quote_manage.AQuoteItem, error)
}

type defaultFilingLookRepository struct {
	db *DBEngine
}

func (r *defaultFilingLookRepository) GetValidProcesses(ctx context.Context) ([]*quote_manage.QuoteProcess, error) {
	q := quote_query.Use(r.db.QuoteManage)
	return q.QuoteProcess.WithContext(ctx).Where(
		q.QuoteProcess.PresentStatus.Neq("已撤回"),
	).Order(q.QuoteProcess.ID.Desc()).Find()
}

func (r *defaultFilingLookRepository) GetNodesByProcessIDs(ctx context.Context, processIDs []int32) ([]*quote_manage.QuoteProcessNode, error) {
	if len(processIDs) == 0 {
		return []*quote_manage.QuoteProcessNode{}, nil
	}
	q := quote_query.Use(r.db.QuoteManage)
	return q.QuoteProcessNode.WithContext(ctx).Where(
		q.QuoteProcessNode.ProcessID.In(processIDs...),
	).Order(q.QuoteProcessNode.ID.Desc()).Find()
}

func (r *defaultFilingLookRepository) GetQuotesByQuoteIDs(ctx context.Context, quoteIDs []int32) ([]*quote_manage.Quote, error) {
	if len(quoteIDs) == 0 {
		return []*quote_manage.Quote{}, nil
	}
	q := quote_query.Use(r.db.QuoteManage)
	return q.Quote.WithContext(ctx).Where(
		q.Quote.ID.In(quoteIDs...),
	).Order(q.Quote.ID.Desc()).Find()
}

func (r *defaultFilingLookRepository) GetQuoteItemsByQuoteIDs(ctx context.Context, quoteIDs []int32) ([]*quote_manage.AQuoteItem, error) {
	if len(quoteIDs) == 0 {
		return []*quote_manage.AQuoteItem{}, nil
	}
	q := quote_query.Use(r.db.QuoteManage)
	return q.AQuoteItem.WithContext(ctx).Where(
		q.AQuoteItem.QuoteID.In(quoteIDs...),
	).Order(q.AQuoteItem.ID.Desc()).Find()
}

type filingLookService struct {
	repo filingLookRepository
}

func NewFilingLookService(db *DBEngine) FilingLookService {
	return &filingLookService{
		repo: &defaultFilingLookRepository{db: db},
	}
}

func NewMockFilingLookService(repo filingLookRepository) FilingLookService {
	return &filingLookService{
		repo: repo,
	}
}

func (s *filingLookService) FilingLook(ctx context.Context) (*FilingLookResult, error) {
	result := &FilingLookResult{
		QuoteProcesses:    make([]*quote_manage.QuoteProcess, 0),
		QuoteProcessNodes: make([]*quote_manage.QuoteProcessNode, 0),
		Quotes:            make([]*quote_manage.Quote, 0),
		QuoteItems:        make([]*quote_manage.AQuoteItem, 0),
	}

	// 1. 查未撤回的quote_process
	processes, err := s.repo.GetValidProcesses(ctx)
	if err != nil {
		return nil, err
	}

	if len(processes) == 0 {
		return result, nil
	}

	result.QuoteProcesses = processes

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

	// 2. 查quote_process_node：quote_process_node.process_id in processIDs
	if len(processIDs) > 0 {
		nodes, err := s.repo.GetNodesByProcessIDs(ctx, processIDs)
		if err != nil {
			return nil, err
		}
		if len(nodes) > 0 {
			result.QuoteProcessNodes = nodes
		}
	}

	// 3. 查quote与quote_item：quote.id in quoteIDs
	if len(quoteIDs) > 0 {
		quotes, err := s.repo.GetQuotesByQuoteIDs(ctx, quoteIDs)
		if err != nil {
			return nil, err
		}
		if len(quotes) > 0 {
			result.Quotes = quotes
		}

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
