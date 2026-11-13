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
	GetNodesByApproverID(ctx context.Context, approverID int32) ([]*quote_manage.QuoteProcessNode, error)
	GetValidProcessesByProcessIDs(ctx context.Context, processIDs []int32) ([]*quote_manage.QuoteProcess, error)
	GetNodesByProcessIDs(ctx context.Context, processIDs []int32) ([]*quote_manage.QuoteProcessNode, error)
	GetQuotesByQuoteIDs(ctx context.Context, quoteIDs []int32) ([]*quote_manage.Quote, error)
	GetQuoteItemsByQuoteIDs(ctx context.Context, quoteIDs []int32) ([]*quote_manage.AQuoteItem, error)
}

type defaultMyApproveQueryRepository struct {
	db *DBEngine
}

func (r *defaultMyApproveQueryRepository) GetNodesByApproverID(ctx context.Context, approverID int32) ([]*quote_manage.QuoteProcessNode, error) {
	q := quote_query.Use(r.db.QuoteManage)
	return q.QuoteProcessNode.WithContext(ctx).Where(
		q.QuoteProcessNode.ApproverID.Eq(approverID),
	).Order(q.QuoteProcessNode.ID.Desc()).Find()
}

func (r *defaultMyApproveQueryRepository) GetValidProcessesByProcessIDs(ctx context.Context, processIDs []int32) ([]*quote_manage.QuoteProcess, error) {
	if len(processIDs) == 0 {
		return []*quote_manage.QuoteProcess{}, nil
	}
	q := quote_query.Use(r.db.QuoteManage)
	return q.QuoteProcess.WithContext(ctx).Where(
		q.QuoteProcess.ID.In(processIDs...),
		q.QuoteProcess.PresentStatus.Neq("已撤回"),
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

	// 1. 用approver_id == user.id查quote_process_node表，得到quote_process_node记录列表，
	//    提取出“process_id列表”
	userNodes, err := s.repo.GetNodesByApproverID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(userNodes) == 0 {
		return result, nil
	}

	processIDMap := make(map[int32]bool)
	processIDs := make([]int32, 0, len(userNodes))
	for _, node := range userNodes {
		if node != nil && node.ProcessID != nil && !processIDMap[*node.ProcessID] {
			processIDMap[*node.ProcessID] = true
			processIDs = append(processIDs, *node.ProcessID)
		}
	}

	if len(processIDs) == 0 {
		return result, nil
	}

	// 2. 用id in “process_id列表” 且 present_status != '已撤回'查quote_process表，
	//    得到【有效quote_process列表】，并提取出“quote_id列表”、“有效quote_process.id列表”
	validProcesses, err := s.repo.GetValidProcessesByProcessIDs(ctx, processIDs)
	if err != nil {
		return nil, err
	}
	if len(validProcesses) == 0 {
		return result, nil
	}

	result.QuoteProcesses = validProcesses
	result.Total = int64(len(validProcesses))

	validProcessIDs := make([]int32, 0, len(validProcesses))
	quoteIDMap := make(map[int32]bool)
	quoteIDs := make([]int32, 0, len(validProcesses))

	for _, p := range validProcesses {
		if p == nil {
			continue
		}
		validProcessIDs = append(validProcessIDs, p.ID)
		if p.QuoteID != nil && !quoteIDMap[*p.QuoteID] {
			quoteIDMap[*p.QuoteID] = true
			quoteIDs = append(quoteIDs, *p.QuoteID)
		}
	}

	// 3. 用process_id in “有效quote_process.id列表”查quote_process_node表，得到【quote_process_node列表】
	//    （为了查到quote_process所有quote_process_node记录）
	if len(validProcessIDs) > 0 {
		nodes, err := s.repo.GetNodesByProcessIDs(ctx, validProcessIDs)
		if err != nil {
			return nil, err
		}
		if len(nodes) > 0 {
			result.QuoteProcessNodes = nodes
		}
	}

	// 4. 用id in “quote_id列表”查quote表，得到【quote列表】
	if len(quoteIDs) > 0 {
		quotes, err := s.repo.GetQuotesByQuoteIDs(ctx, quoteIDs)
		if err != nil {
			return nil, err
		}
		if len(quotes) > 0 {
			result.Quotes = quotes
		}

		// 5. 用quote_id in “quote_id列表”查A_quote_item表，得到【A_quote_item列表】
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
