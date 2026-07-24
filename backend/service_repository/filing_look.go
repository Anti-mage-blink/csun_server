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
	GetAllProcesses(ctx context.Context) ([]*quote_manage.QuoteProcess, error)
	GetAllNodes(ctx context.Context) ([]*quote_manage.QuoteProcessNode, error)
	GetAllQuotes(ctx context.Context) ([]*quote_manage.Quote, error)
	GetAllQuoteItems(ctx context.Context) ([]*quote_manage.AQuoteItem, error)
}

type defaultFilingLookRepository struct {
	db *DBEngine
}

func (r *defaultFilingLookRepository) GetAllProcesses(ctx context.Context) ([]*quote_manage.QuoteProcess, error) {
	q := quote_query.Use(r.db.QuoteManage)
	return q.QuoteProcess.WithContext(ctx).Order(q.QuoteProcess.ID.Desc()).Find()
}

func (r *defaultFilingLookRepository) GetAllNodes(ctx context.Context) ([]*quote_manage.QuoteProcessNode, error) {
	q := quote_query.Use(r.db.QuoteManage)
	return q.QuoteProcessNode.WithContext(ctx).Order(q.QuoteProcessNode.ID.Desc()).Find()
}

func (r *defaultFilingLookRepository) GetAllQuotes(ctx context.Context) ([]*quote_manage.Quote, error) {
	q := quote_query.Use(r.db.QuoteManage)
	return q.Quote.WithContext(ctx).Order(q.Quote.ID.Desc()).Find()
}

func (r *defaultFilingLookRepository) GetAllQuoteItems(ctx context.Context) ([]*quote_manage.AQuoteItem, error) {
	q := quote_query.Use(r.db.QuoteManage)
	return q.AQuoteItem.WithContext(ctx).Order(q.AQuoteItem.ID.Desc()).Find()
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

	processes, err := s.repo.GetAllProcesses(ctx)
	if err != nil {
		return nil, err
	}
	if len(processes) > 0 {
		result.QuoteProcesses = processes
	}

	nodes, err := s.repo.GetAllNodes(ctx)
	if err != nil {
		return nil, err
	}
	if len(nodes) > 0 {
		result.QuoteProcessNodes = nodes
	}

	quotes, err := s.repo.GetAllQuotes(ctx)
	if err != nil {
		return nil, err
	}
	if len(quotes) > 0 {
		result.Quotes = quotes
	}

	items, err := s.repo.GetAllQuoteItems(ctx)
	if err != nil {
		return nil, err
	}
	if len(items) > 0 {
		result.QuoteItems = items
	}

	return result, nil
}
