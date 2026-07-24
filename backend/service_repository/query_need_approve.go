package service_repository

import (
	"context"
	"errors"

	"csun_server-backend/dao/model/quote_manage"
	quote_query "csun_server-backend/dao/query/quote_manage"

	"gorm.io/gorm"
)

// QueryNeedApproveResult 定义待审批查询结果的数据结构
type QueryNeedApproveResult struct {
	Total             int64                            `json:"total"`
	QuoteProcesses    []*quote_manage.QuoteProcess     `json:"quote_processes"`
	QuoteProcessNodes []*quote_manage.QuoteProcessNode `json:"quote_process_nodes"`
	Quotes            []*quote_manage.Quote            `json:"quotes"`
	QuoteItems        []*quote_manage.AQuoteItem       `json:"quote_items"`
}

// QueryNeedApproveService 待审批查询服务接口
type QueryNeedApproveService interface {
	QueryNeedApprove(ctx context.Context, userID int32) (*QueryNeedApproveResult, error)
}

type queryNeedApproveRepository interface {
	GetPendingNodesByEmployeeID(ctx context.Context, employeeID int32) ([]*quote_manage.QuoteProcessNode, error)
	GetProcessByID(ctx context.Context, processID int32) (*quote_manage.QuoteProcess, error)
	GetNodesByProcessID(ctx context.Context, processID int32) ([]*quote_manage.QuoteProcessNode, error)
	GetQuoteByID(ctx context.Context, quoteID int32) (*quote_manage.Quote, error)
	GetQuoteItemsByQuoteID(ctx context.Context, quoteID int32) ([]*quote_manage.AQuoteItem, error)
}

type defaultQueryNeedApproveRepository struct {
	db *DBEngine
}

func (r *defaultQueryNeedApproveRepository) GetPendingNodesByEmployeeID(ctx context.Context, employeeID int32) ([]*quote_manage.QuoteProcessNode, error) {
	q := quote_query.Use(r.db.QuoteManage)
	return q.QuoteProcessNode.WithContext(ctx).Where(
		q.QuoteProcessNode.ApproveEmployeeID.Eq(employeeID),
		q.QuoteProcessNode.Status.Eq("待审批"),
	).Find()
}

func (r *defaultQueryNeedApproveRepository) GetProcessByID(ctx context.Context, processID int32) (*quote_manage.QuoteProcess, error) {
	q := quote_query.Use(r.db.QuoteManage)
	process, err := q.QuoteProcess.WithContext(ctx).Where(
		q.QuoteProcess.ID.Eq(processID),
	).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return process, nil
}

func (r *defaultQueryNeedApproveRepository) GetNodesByProcessID(ctx context.Context, processID int32) ([]*quote_manage.QuoteProcessNode, error) {
	q := quote_query.Use(r.db.QuoteManage)
	return q.QuoteProcessNode.WithContext(ctx).Where(
		q.QuoteProcessNode.ProcessID.Eq(processID),
	).Find()
}

func (r *defaultQueryNeedApproveRepository) GetQuoteByID(ctx context.Context, quoteID int32) (*quote_manage.Quote, error) {
	q := quote_query.Use(r.db.QuoteManage)
	quote, err := q.Quote.WithContext(ctx).Where(
		q.Quote.ID.Eq(quoteID),
	).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return quote, nil
}

func (r *defaultQueryNeedApproveRepository) GetQuoteItemsByQuoteID(ctx context.Context, quoteID int32) ([]*quote_manage.AQuoteItem, error) {
	q := quote_query.Use(r.db.QuoteManage)
	return q.AQuoteItem.WithContext(ctx).Where(
		q.AQuoteItem.QuoteID.Eq(quoteID),
	).Find()
}

type queryNeedApproveService struct {
	repo queryNeedApproveRepository
}

func NewQueryNeedApproveService(db *DBEngine) QueryNeedApproveService {
	return &queryNeedApproveService{
		repo: &defaultQueryNeedApproveRepository{db: db},
	}
}

func NewMockQueryNeedApproveService(repo queryNeedApproveRepository) QueryNeedApproveService {
	return &queryNeedApproveService{
		repo: repo,
	}
}

func (s *queryNeedApproveService) QueryNeedApprove(ctx context.Context, userID int32) (*QueryNeedApproveResult, error) {
	// 初始化返回结构体，默认给空切片，避免前端接收到 null
	result := &QueryNeedApproveResult{
		Total:             0,
		QuoteProcesses:    make([]*quote_manage.QuoteProcess, 0),
		QuoteProcessNodes: make([]*quote_manage.QuoteProcessNode, 0),
		Quotes:            make([]*quote_manage.Quote, 0),
		QuoteItems:        make([]*quote_manage.AQuoteItem, 0),
	}

	// 1. 查询待审批的审批节点记录
	pendingNodes, err := s.repo.GetPendingNodesByEmployeeID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(pendingNodes) == 0 {
		return result, nil
	}

	result.Total = int64(len(pendingNodes))

	// 用于去重并保持结果顺序的 map set
	processSet := make(map[int32]bool)
	nodeSet := make(map[int32]bool)
	quoteSet := make(map[int32]bool)
	itemSet := make(map[int32]bool)

	// 2. 遍历查到的待审批节点记录
	for _, node := range pendingNodes {
		if node.ProcessID == nil {
			continue
		}
		processID := *node.ProcessID

		// 用 process_id 查 quote_process 得到记录
		process, err := s.repo.GetProcessByID(ctx, processID)
		if err != nil {
			return nil, err
		}
		if process != nil {
			if !processSet[process.ID] {
				processSet[process.ID] = true
				result.QuoteProcesses = append(result.QuoteProcesses, process)
			}
		}

		// 用 process_id 查所有同一 process_id 的节点
		nodes, err := s.repo.GetNodesByProcessID(ctx, processID)
		if err != nil {
			return nil, err
		}
		for _, n := range nodes {
			if n != nil {
				if !nodeSet[n.ID] {
					nodeSet[n.ID] = true
					result.QuoteProcessNodes = append(result.QuoteProcessNodes, n)
				}
			}
		}
	}

	// 遍历“quote_process记录列表” 用 quote_id 查询 quote 表
	for _, process := range result.QuoteProcesses {
		if process.QuoteID == nil {
			continue
		}
		quoteID := *process.QuoteID

		quote, err := s.repo.GetQuoteByID(ctx, quoteID)
		if err != nil {
			return nil, err
		}
		if quote != nil {
			if !quoteSet[quote.ID] {
				quoteSet[quote.ID] = true
				result.Quotes = append(result.Quotes, quote)
			}
		}
	}

	// 遍历“quote记录列表” 用 quote_id 查询 A_quote_item 表
	for _, q := range result.Quotes {
		quoteID := q.ID

		items, err := s.repo.GetQuoteItemsByQuoteID(ctx, quoteID)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if item != nil {
				if !itemSet[item.ID] {
					itemSet[item.ID] = true
					result.QuoteItems = append(result.QuoteItems, item)
				}
			}
		}
	}

	return result, nil
}
