package service_repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"csun_server-backend/dao/model/quote_manage"
	quote_query "csun_server-backend/dao/query/quote_manage"
	"csun_server-backend/utils"

	"gorm.io/gorm"
)

// ApproveHandleRequestParams 业务逻辑接收的入参参数
type ApproveHandleRequestParams struct {
	Action    string // "approve" 或 "reject"
	NodeID    int32
	ProcessID int32
	Comment   string
	UserID    int32
	UserName  string
	UserRole  string
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
		qQuote := quote_query.Use(tx)

		// 1. 查询并验证节点是否存在，且属于该审批流
		node, err := qQuote.QuoteProcessNode.WithContext(ctx).Where(qQuote.QuoteProcessNode.ID.Eq(params.NodeID)).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("审批流节点记录未找到")
			}
			return err
		}

		if node.ProcessID == nil || *node.ProcessID != params.ProcessID {
			return errors.New("审批流节点与审批流不匹配")
		}

		// 2. 统一生成当下时间
		now := time.Now()

		// 3. 确定状态转换: 已同意 / 已拒绝
		var statusStr string
		switch params.Action {
		case "approve":
			statusStr = "已同意"
		case "reject":
			statusStr = "已拒绝"
		default:
			return errors.New("不支持的审批操作")
		}

		// 4. 更新当前节点 quote_process_node 记录的字段
		// status: 已同意/已拒绝
		// approver_id: user.id
		// approver_name: user.name
		// comment: 前端传来的 comment
		// approve_at: 当下时间
		node.Status = &statusStr
		if params.UserID > 0 {
			node.ApproverID = &params.UserID
		}
		if params.UserName != "" {
			node.ApproverName = &params.UserName
		}
		node.ApproveComment = &params.Comment
		node.ApproveAt = &now

		if _, err := qQuote.QuoteProcessNode.WithContext(ctx).Where(qQuote.QuoteProcessNode.ID.Eq(params.NodeID)).Updates(node); err != nil {
			return err
		}

		// 5. 更新当前流程 quote_process 记录的字段
		// present_status: 已同意/已拒绝
		process, err := qQuote.QuoteProcess.WithContext(ctx).Where(qQuote.QuoteProcess.ID.Eq(params.ProcessID)).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("审批流记录未找到")
			}
			return err
		}

		process.PresentStatus = &statusStr
		process.UpdatedAt = &now
		if _, err := qQuote.QuoteProcess.WithContext(ctx).Where(qQuote.QuoteProcess.ID.Eq(params.ProcessID)).Updates(process); err != nil {
			return err
		}

		// 6. （意图）新创下一节点（仅在同意 approve 时操作）
		if params.Action == "approve" {
			var condParamName string
			var condParamValue interface{}

			if params.UserRole == "领导小组副组长" {
				condParamName = "quote_float_rate"
				if process.QuoteID != nil {
					item, err := qQuote.AQuoteItem.WithContext(ctx).Where(qQuote.AQuoteItem.QuoteID.Eq(*process.QuoteID)).First()
					if err == nil && item != nil && item.QuoteFloatRate != nil {
						condParamValue = *item.QuoteFloatRate
					}
				}
			}

			//  调用“查询下一节点编号”函数
			nextNode, err := utils.QueryNextNodeNum(ctx, r.db.General, process.PresentNodeNum, condParamName, condParamValue)
			if err != nil {
				return fmt.Errorf("查询下一节点编号失败: %w", err)
			}

			if nextNode != nil {
				// if next_node 不为空:
				// 查到人：如果 next_node.role 不为空，去 quote_manage.employee_role 查到记录 employee_role
				var approveEmpID *int32
				var approveEmpName *string
				if nextNode.Role != nil && *nextNode.Role != "" {
					empRole, err := qQuote.EmployeeRole.WithContext(ctx).
						Where(qQuote.EmployeeRole.Role.Eq(*nextNode.Role)).
						First()
					if err == nil && empRole != nil {
						approveEmpID = empRole.EmployeeID
						approveEmpName = empRole.EmployeeName
					}
				}
				// 打印approveEmpID和approveEmpName
				fmt.Printf("Approve Employee ID: %v, Name: %v\n", approveEmpID, approveEmpName)

				// 构造节点实例记录 quote_process_node
				var nextSeqNum int32 = 1
				if process.PresentSeqNum != nil {
					nextSeqNum = *process.PresentSeqNum + 1
				}
				statusPending := "待审批"

				presentNode := &quote_manage.QuoteProcessNode{
					ProcessID:    &process.ID,
					SeqNum:       &nextSeqNum,
					NodeNum:      nextNode.NodeNum,
					Name:         nextNode.Name,
					Role:         nextNode.Role,
					ApproverID:   approveEmpID,
					ApproverName: approveEmpName,
					Status:       &statusPending,
					CreatedAt:    &now,
				}

				if err := qQuote.QuoteProcessNode.WithContext(ctx).Create(presentNode); err != nil {
					return fmt.Errorf("写入下一节点实例失败: %w", err)
				}

				// 更新当前 quote_process 记录
				process.PresentStatus = &statusPending
				process.PresentSeqNum = &nextSeqNum
				process.PresentNodeNum = nextNode.NodeNum
				process.PresentNodeID = &presentNode.ID
				process.PresentNodeName = presentNode.Name
				process.PresentApproverID = approveEmpID
				process.PresentApproverName = approveEmpName
				process.UpdatedAt = &now

				if err := qQuote.QuoteProcess.WithContext(ctx).Save(process); err != nil {
					return fmt.Errorf("更新审批流实例记录失败: %w", err)
				}
			} else {
				// else（即 next_node 为空）
				// present_status: "通过/拒绝（已结束）"，此处已为 "已同意"，更新 updated_at
				process.UpdatedAt = &now
				if err := qQuote.QuoteProcess.WithContext(ctx).Save(process); err != nil {
					return fmt.Errorf("更新审批流实例记录失败: %w", err)
				}
			}
		}

		return nil
	})
}

type approveHandleService struct {
	repo approveHandleRepository
}

// NewApproveHandleService创建默认审批处理服务
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
