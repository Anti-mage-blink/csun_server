package service_repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"csun_server-backend/dao/model/general"
	"csun_server-backend/dao/model/quote_manage"
	general_query "csun_server-backend/dao/query/general"
	quote_query "csun_server-backend/dao/query/quote_manage"
	"csun_server-backend/utils"

	"gorm.io/gorm"
)

// CreateQuoteData 发起报价单返回给前端的全量数据
type CreateQuoteData struct {
	Quote        *quote_manage.Quote         `json:"quote"`
	QuoteItem    *quote_manage.AQuoteItem    `json:"quote_item"`
	Customers    []*general.Customer         `json:"customers"`
	ProductSpecs []*quote_manage.ProductSpec `json:"product_specs"`
}

type CreateQuoteService interface {
	PrepareCreateQuote(ctx context.Context) (*CreateQuoteData, error)
	SubmitQuote(ctx context.Context, quote *quote_manage.Quote, items []*quote_manage.AQuoteItem, userID int32, userName string) error
}

type createQuoteRepository interface {
	GetMaxQuoteCodeByPrefix(ctx context.Context, prefix string) (*string, error)
	GetAllCustomers(ctx context.Context) ([]*general.Customer, error)
	GetAllProductSpecs(ctx context.Context) ([]*quote_manage.ProductSpec, error)
	SaveQuoteWithItems(ctx context.Context, quote *quote_manage.Quote, items []*quote_manage.AQuoteItem, userID int32, userName string) error
}

type defaultCreateQuoteRepository struct {
	db *DBEngine
}

func (r *defaultCreateQuoteRepository) GetMaxQuoteCodeByPrefix(ctx context.Context, prefix string) (*string, error) {
	q := quote_query.Use(r.db.QuoteManage)
	quote, err := q.Quote.WithContext(ctx).
		Where(q.Quote.QuoteCode.Like(prefix + "%")).
		Order(q.Quote.QuoteCode.Desc()).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return quote.QuoteCode, nil
}

func (r *defaultCreateQuoteRepository) GetAllCustomers(ctx context.Context) ([]*general.Customer, error) {
	q := general_query.Use(r.db.General)
	return q.Customer.WithContext(ctx).Find()
}

func (r *defaultCreateQuoteRepository) GetAllProductSpecs(ctx context.Context) ([]*quote_manage.ProductSpec, error) {
	q := quote_query.Use(r.db.QuoteManage)
	return q.ProductSpec.WithContext(ctx).Find()
}

func (r *defaultCreateQuoteRepository) SaveQuoteWithItems(ctx context.Context, quote *quote_manage.Quote, items []*quote_manage.AQuoteItem, userID int32, userName string) error {
	return r.db.QuoteManage.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		qQuote := quote_query.Use(tx)
		qGen := general_query.Use(r.db.General)

		// 0. 在数据库 general 数据表 customer 中用前端传来的三字段 company_name、contact_name、contact_title（与关系，同时满足）查询记录
		var compName, contName, contTitle string
		if quote.CustomerName != nil {
			compName = *quote.CustomerName
		}
		if quote.ContactName != nil {
			contName = *quote.ContactName
		}
		if quote.ContactTitle != nil {
			contTitle = *quote.ContactTitle
		}

		if compName != "" || contName != "" || contTitle != "" {
			cust, err := qGen.Customer.WithContext(ctx).
				Where(
					qGen.Customer.CompanyName.Eq(compName),
					qGen.Customer.ContactName.Eq(contName),
					qGen.Customer.ContactTitle.Eq(contTitle),
				).First()

			if errors.Is(err, gorm.ErrRecordNotFound) || cust == nil {
				// 若查不到，则给该表新增记录
				newCust := &general.Customer{
					CompanyName:  quote.CustomerName,
					ContactName:  quote.ContactName,
					ContactTitle: quote.ContactTitle,
				}
				if err := qGen.Customer.WithContext(ctx).Create(newCust); err != nil {
					return err
				}
				quote.CustomerID = &newCust.ID // 直接更新为新建客户记录的自增 ID
			} else if err != nil {
				return err
			} else {
				// 若能查到，同步更新报价单绑定的客户 ID
				quote.CustomerID = &cust.ID
			}
		}

		// 1. 先将报价单写入数据库：quote_manage数据库quote数据表，拿到写入时自增后的id
		if err := qQuote.Quote.WithContext(ctx).Create(quote); err != nil {
			return err
		}

		// 2. 再将这个id填入到每个报价明细单结构体quote_id中，填写Is_below_floor_price， 再写入quote_manage数据库A_quote_item数据表
		for _, item := range items {
			item.QuoteID = &quote.ID

			// 判断单价（quote_unit_price）是否小于该产品 product_spec 记录的底价（floor_price）：若小于则为 true，否则为 false
			isBelow := false
			if item.ProductSpecID != nil && item.QuoteUnitPrice != nil {
				spec, err := qQuote.ProductSpec.WithContext(ctx).Where(qQuote.ProductSpec.ID.Eq(*item.ProductSpecID)).First()
				if err == nil && spec != nil && spec.FloorPrice != nil {
					if *item.QuoteUnitPrice < *spec.FloorPrice {
						isBelow = true
					} else {
						isBelow = false
					}
				}
			}
			item.IsBelowFloorPrice = &isBelow

			if err := qQuote.AQuoteItem.WithContext(ctx).Create(item); err != nil {
				return err
			}
		}

		// 3. 新创建报价审批流记录：即quote_manage.quote_process
		now := time.Now()
		defaultStatus := "待审批"
		process := &quote_manage.QuoteProcess{
			QuoteID:       &quote.ID,
			CreatorID:     &userID,
			CreatorName:   &userName,
			PresentStatus: &defaultStatus,
			CreatedAt:     &now,
			UpdatedAt:     &now,
		}
		if err := qQuote.QuoteProcess.WithContext(ctx).Create(process); err != nil {
			return err
		}

		// 4. （意图）新创开始“发起报价单”节点：
		// 调用“查询下一节点编号”函数
		nextStartNode, err := utils.QueryNextNodeNum(ctx, r.db.General, process.PresentNodeNum, "", nil)
		if err != nil {
			return fmt.Errorf("查询开始节点失败: %w", err)
		}
		if nextStartNode == nil {
			return fmt.Errorf("未找到开始流程节点")
		}

		// 构造节点实例记录 quote_process_node
		var seqNum1 int32 = 1
		statusPassed := "已通过"
		startNodeInstance := &quote_manage.QuoteProcessNode{
			ProcessID: &process.ID,
			SeqNum:    &seqNum1,
			NodeNum:   nextStartNode.NodeNum,
			Name:      nextStartNode.Name,
			Role:      nextStartNode.Role,
			Status:    &statusPassed,
			CreatedAt: &now,
			ApproveAt: &now,
		}
		if err := qQuote.QuoteProcessNode.WithContext(ctx).Create(startNodeInstance); err != nil {
			return fmt.Errorf("写入开始节点实例失败: %w", err)
		}

		// 更新当前处理的 quote_process 记录
		process.QuoteID = &quote.ID
		process.CreatorID = &userID
		process.CreatorName = &userName
		process.PresentSeqNum = &seqNum1
		process.PresentNodeNum = nextStartNode.NodeNum
		process.CreatedAt = &now

		if err := qQuote.QuoteProcess.WithContext(ctx).Save(process); err != nil {
			return fmt.Errorf("更新审批流实例记录失败: %w", err)
		}

		// 5. （意图）新创“工作小组组长”节点：
		// 调用“查询下一节点编号”函数
		var mainCategory string
		if len(items) > 0 && items[0] != nil && items[0].ProductCategoryName != nil {
			mainCategory = *items[0].ProductCategoryName
		}

		nextGroupNode, err := utils.QueryNextNodeNum(ctx, r.db.General, process.PresentNodeNum, "main_category", mainCategory)
		if err != nil {
			return fmt.Errorf("查询下一审批节点失败: %w", err)
		}
		if nextGroupNode == nil {
			return fmt.Errorf("未找到下一审批节点")
		}

		// 查到人：如果 next_node.role 不为空，去 quote_manage.employee_role 查到记录 employee_role
		var approveEmpID *int32
		var approveEmpName *string
		if nextGroupNode.Role != nil && *nextGroupNode.Role != "" {
			empRole, err := qQuote.EmployeeRole.WithContext(ctx).
				Where(qQuote.EmployeeRole.Role.Eq(*nextGroupNode.Role)).
				First()
			if err == nil && empRole != nil {
				approveEmpID = empRole.EmployeeID
				approveEmpName = empRole.EmployeeName
			}
		}

		// 构造节点实例记录 quote_process_node
		nextSeqNum := *process.PresentSeqNum + 1
		statusPending := "待审批"

		groupNodeInstance := &quote_manage.QuoteProcessNode{
			ProcessID:    &process.ID,
			SeqNum:       &nextSeqNum,
			NodeNum:      nextGroupNode.NodeNum,
			Name:         nextGroupNode.Name,
			Role:         nextGroupNode.Role,
			ApproverID:   approveEmpID,
			ApproverName: approveEmpName,
			Status:       &statusPending,
			CreatedAt:    &now,
		}
		if err := qQuote.QuoteProcessNode.WithContext(ctx).Create(groupNodeInstance); err != nil {
			return fmt.Errorf("写入审批节点实例失败: %w", err)
		}

		// 更新当前处理的 quote_process 记录
		presentNodeID := groupNodeInstance.ID
		process.QuoteID = &quote.ID
		process.CreatorID = &userID
		process.CreatorName = &userName
		process.PresentStatus = &statusPending
		process.PresentSeqNum = &nextSeqNum
		process.PresentNodeNum = nextGroupNode.NodeNum
		process.PresentNodeID = &presentNodeID
		process.PresentNodeName = groupNodeInstance.Name
		process.PresentApproverID = approveEmpID
		process.PresentApproverName = approveEmpName
		process.UpdatedAt = &now

		if err := qQuote.QuoteProcess.WithContext(ctx).Save(process); err != nil {
			return fmt.Errorf("更新审批流实例记录失败: %w", err)
		}

		return nil
	})
}

type createQuoteService struct {
	repo createQuoteRepository
}

func NewCreateQuoteService(db *DBEngine) CreateQuoteService {
	return &createQuoteService{
		repo: &defaultCreateQuoteRepository{db: db},
	}
}

func NewMockCreateQuoteService(repo createQuoteRepository) CreateQuoteService {
	return &createQuoteService{
		repo: repo,
	}
}

func (s *createQuoteService) PrepareCreateQuote(ctx context.Context) (*CreateQuoteData, error) {
	// 1. 生成报价单编号：前缀 "BJ-{今日年月日}-"
	today := time.Now().Format("20060102")
	prefix := fmt.Sprintf("BJ-%s-", today)

	maxCode, err := s.repo.GetMaxQuoteCodeByPrefix(ctx, prefix)
	if err != nil {
		return nil, err
	}
	quoteCode := generateQuoteCode(prefix, maxCode)

	// 2. 获取全量数据表，方便前端使用
	customers, err := s.repo.GetAllCustomers(ctx)
	if err != nil {
		return nil, err
	}
	productSpecs, err := s.repo.GetAllProductSpecs(ctx)
	if err != nil {
		return nil, err
	}

	return &CreateQuoteData{
		Quote: &quote_manage.Quote{
			QuoteCode: &quoteCode,
		},
		QuoteItem:    &quote_manage.AQuoteItem{},
		Customers:    customers,
		ProductSpecs: productSpecs,
	}, nil
}

func generateQuoteCode(prefix string, maxCode *string) string {
	const defaultSeq = "001"
	if maxCode == nil {
		return prefix + defaultSeq
	}
	seqStr := (*maxCode)[len(prefix):]
	seq, err := strconv.Atoi(seqStr)
	if err != nil {
		return prefix + defaultSeq
	}
	return fmt.Sprintf("%s%03d", prefix, seq+1)
}

func (s *createQuoteService) SubmitQuote(ctx context.Context, quote *quote_manage.Quote, items []*quote_manage.AQuoteItem, userID int32, userName string) error {
	if quote == nil {
		return fmt.Errorf("quote cannot be nil")
	}
	if len(items) == 0 {
		return fmt.Errorf("quote items cannot be empty")
	}
	if quote.PayWay == nil || *quote.PayWay == "" {
		return fmt.Errorf("付款方式(pay_way)不能为空")
	}
	if quote.CreditPeriod == nil || *quote.CreditPeriod == "" {
		return fmt.Errorf("账期(credit_period)不能为空")
	}
	return s.repo.SaveQuoteWithItems(ctx, quote, items, userID, userName)
}
