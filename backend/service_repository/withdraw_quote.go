package service_repository

import (
	"context"
	"errors"
	"time"

	"csun_server-backend/dao/model/quote_manage"
	general_query "csun_server-backend/dao/query/general"
	quote_query "csun_server-backend/dao/query/quote_manage"

	"gorm.io/gorm"
)

// WithdrawQuoteRequestParams 撤回报价单业务入参
type WithdrawQuoteRequestParams struct {
	ProcessID int32  // 审批流记录ID (quote_process.id)
	UserID    int32  // 用户ID
	UserName  string // 用户姓名
}

// WithdrawQuoteService 撤回报价单服务接口
type WithdrawQuoteService interface {
	WithdrawQuote(ctx context.Context, params *WithdrawQuoteRequestParams) error
}

type withdrawQuoteRepository interface {
	WithdrawQuote(ctx context.Context, params *WithdrawQuoteRequestParams) error
}

type defaultWithdrawQuoteRepository struct {
	db *DBEngine
}

func (r *defaultWithdrawQuoteRepository) WithdrawQuote(ctx context.Context, params *WithdrawQuoteRequestParams) error {
	// 如果 params.UserName 为空，尝试从 general.employee 数据表中查询用户姓名
	if params.UserName == "" && params.UserID > 0 {
		qGen := general_query.Use(r.db.General)
		emp, err := qGen.Employee.WithContext(ctx).Where(qGen.Employee.ID.Eq(params.UserID)).First()
		if err == nil && emp != nil && emp.Name != nil {
			params.UserName = *emp.Name
		}
	}

	return r.db.QuoteManage.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		q := quote_query.Use(tx)

		// 1. 查询该审批流下的所有节点，查出 seq_num 最大的那条记录（相同 seq_num 时按 ID 降序）
		nodes, err := q.QuoteProcessNode.WithContext(ctx).
			Where(q.QuoteProcessNode.ProcessID.Eq(params.ProcessID)).
			Order(q.QuoteProcessNode.SeqNum.Desc(), q.QuoteProcessNode.ID.Desc()).
			Find()
		if err != nil {
			return err
		}
		if len(nodes) == 0 {
			return errors.New("未找到需要撤回的审批节点记录")
		}

		// 2. 删掉该 seq_num 最大的节点记录
		lastNode := nodes[0]
		if _, err := q.QuoteProcessNode.WithContext(ctx).
			Where(q.QuoteProcessNode.ID.Eq(lastNode.ID)).
			Delete(lastNode); err != nil {
			return err
		}

		// 3. 新增一个节点记录：quote_process_node
		nowStr := time.Now().Format("2006-01-02 15:04:05")
		seqNum := int32(2)
		nodeName := "撤回报价单"
		statusPass := "已通过"

		newNode := &quote_manage.QuoteProcessNode{
			ProcessID:           &params.ProcessID,
			SeqNum:              &seqNum,
			Name:                &nodeName,
			ApproveEmployeeID:   &params.UserID,
			ApproveEmployeeName: &params.UserName,
			Status:              &statusPass,
			CreatedAt:           &nowStr,
			ApproveAt:           &nowStr,
		}

		if err := q.QuoteProcessNode.WithContext(ctx).Create(newNode); err != nil {
			return err
		}

		// 4. 更新审批流记录：quote_process_123
		process, err := q.QuoteProcess.WithContext(ctx).
			Where(q.QuoteProcess.ID.Eq(params.ProcessID)).
			First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("审批流记录未找到")
			}
			return err
		}

		presentStatus := "已撤回"
		process.PresentStatus = &presentStatus
		process.PresentNodeID = &newNode.ID
		process.PresentNodeName = &nodeName

		if _, err := q.QuoteProcess.WithContext(ctx).
			Where(q.QuoteProcess.ID.Eq(params.ProcessID)).
			Updates(process); err != nil {
			return err
		}

		return nil
	})
}

type withdrawQuoteService struct {
	repo withdrawQuoteRepository
}

// NewWithdrawQuoteService 创建默认撤回报价单服务
func NewWithdrawQuoteService(db *DBEngine) WithdrawQuoteService {
	return &withdrawQuoteService{
		repo: &defaultWithdrawQuoteRepository{db: db},
	}
}

// NewMockWithdrawQuoteService 创建 mock 撤回报价单服务用于测试
func NewMockWithdrawQuoteService(repo withdrawQuoteRepository) WithdrawQuoteService {
	return &withdrawQuoteService{
		repo: repo,
	}
}

func (s *withdrawQuoteService) WithdrawQuote(ctx context.Context, params *WithdrawQuoteRequestParams) error {
	if params == nil {
		return errors.New("params 不能为空")
	}
	if params.ProcessID <= 0 {
		return errors.New("process_id 必须大于 0")
	}
	if params.UserID <= 0 {
		return errors.New("user_id 必须大于 0")
	}

	return s.repo.WithdrawQuote(ctx, params)
}
