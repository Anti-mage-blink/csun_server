package service_repository

import (
	"context"
	"errors"
	"time"

	quote_query "csun_server-backend/dao/query/quote_manage"

	"gorm.io/gorm"
)

// ApproveHandleRequestParams 业务逻辑接收的入参参数
type ApproveHandleRequestParams struct {
	Action    string // "approve" 或 "reject"
	NodeID    int32
	ProcessID int32
	Comment   string
}

// ApproveHandleService 审批处理服务接口
type ApproveHandleService interface {
	ApproveHandle(ctx context.Context, params *ApproveHandleRequestParams) error
}

type approveHandleRepository interface {
	UpdateNodeAndProcessStatus(ctx context.Context, params *ApproveHandleRequestParams) error
}

type defaultApproveHandleRepository struct {
	db *DBEngine
}

func (r *defaultApproveHandleRepository) UpdateNodeAndProcessStatus(ctx context.Context, params *ApproveHandleRequestParams) error {
	return r.db.QuoteManage.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		q := quote_query.Use(tx)

		// 1. 查询并验证节点是否存在，且属于该审批流
		node, err := q.QuoteProcessNode.WithContext(ctx).Where(q.QuoteProcessNode.ID.Eq(params.NodeID)).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("审批流节点记录未找到")
			}
			return err
		}

		if node.ProcessID == nil || *node.ProcessID != params.ProcessID {
			return errors.New("审批流节点与审批流不匹配")
		}

		// 2. 确定状态转换
		var statusStr string
		if params.Action == "approve" {
			statusStr = "已通过"
		} else if params.Action == "reject" {
			statusStr = "已拒绝"
		} else {
			return errors.New("不支持的审批操作")
		}

		// 3. 更新节点记录
		nowStr := time.Now().Format("2006-01-02 15:04:05")
		node.Status = &statusStr
		node.ApproveComment = &params.Comment
		node.ApproveAt = &nowStr

		if _, err := q.QuoteProcessNode.WithContext(ctx).Where(q.QuoteProcessNode.ID.Eq(params.NodeID)).Updates(node); err != nil {
			return err
		}

		// 4. 更新审批流记录
		process, err := q.QuoteProcess.WithContext(ctx).Where(q.QuoteProcess.ID.Eq(params.ProcessID)).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("审批流记录未找到")
			}
			return err
		}

		process.PresentStatus = &statusStr
		if _, err := q.QuoteProcess.WithContext(ctx).Where(q.QuoteProcess.ID.Eq(params.ProcessID)).Updates(process); err != nil {
			return err
		}

		return nil
	})
}

type approveHandleService struct {
	repo approveHandleRepository
}

// NewApproveHandleService 创建默认审批处理服务
func NewApproveHandleService(db *DBEngine) ApproveHandleService {
	return &approveHandleService{
		repo: &defaultApproveHandleRepository{db: db},
	}
}

// NewMockApproveHandleService 创建 mock 审批处理服务用于测试
func NewMockApproveHandleService(repo approveHandleRepository) ApproveHandleService {
	return &approveHandleService{
		repo: repo,
	}
}

func (s *approveHandleService) ApproveHandle(ctx context.Context, params *ApproveHandleRequestParams) error {
	if params == nil {
		return errors.New("params 不能为空")
	}
	if params.Action != "approve" && params.Action != "reject" {
		return errors.New("action 参数不正确，必须为 'approve' 或 'reject'")
	}
	if params.NodeID <= 0 {
		return errors.New("node_id 必须大于 0")
	}
	if params.ProcessID <= 0 {
		return errors.New("process_id 必须大于 0")
	}
	return s.repo.UpdateNodeAndProcessStatus(ctx, params)
}
