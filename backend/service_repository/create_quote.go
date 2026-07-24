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

	"gorm.io/gorm"
)

// CreateQuoteData 新建报价单返回给前端的全量数据
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
		qGen := general_query.Use(tx)

		// 1. 先将报价单写入数据库：quote_manage数据库quote数据表，拿到写入时自增后的id
		if err := qQuote.Quote.WithContext(ctx).Create(quote); err != nil {
			return err
		}

		// 2. 再将这个id填入到每个报价明细单结构体quote_id中，再写入quote_manage数据库A_quote_item数据表
		for _, item := range items {
			item.QuoteID = &quote.ID
			if err := qQuote.AQuoteItem.WithContext(ctx).Create(item); err != nil {
				return err
			}
		}

		// 3. 新创建报价审批流记录：即quote_manage.quote_process
		createdAtStr := time.Now().Format("2006-01-02 15:04:05")
		process := &quote_manage.QuoteProcess{
			QuoteID:            &quote.ID,
			CreateEmployeeID:   &userID,
			CreateEmployeeName: &userName,
			CreatedAt:          &createdAtStr,
		}
		if err := qQuote.QuoteProcess.WithContext(ctx).Create(process); err != nil {
			return err
		}

		// 4. 新创建报价审批流节点记录：即quote_manage.quote_process_node
		node1Name := "发起报价单"
		node1Status := "已通过"
		node1Comment := "无"
		seqNum1 := int32(1)
		node1 := &quote_manage.QuoteProcessNode{
			ProcessID:      &process.ID,
			SeqNum:         &seqNum1,
			Name:           &node1Name,
			Status:         &node1Status,
			ApproveComment: &node1Comment,
			CreatedAt:      &createdAtStr,
			ApproveAt:      &createdAtStr,
		}
		if err := qQuote.QuoteProcessNode.WithContext(ctx).Create(node1); err != nil {
			return err
		}

		// 5. 再根据浮动比例条件判断（5%和10%），去查对应角色的人（quote_manage数据库的employee_role表）
		var targetEmployeeID int32
		var targetEmployeeName string

		// 默认
		targetEmployeeID = 81
		targetEmployeeName = "李永泉"

		if len(items) > 0 && items[0] != nil {
			floatRate := 0.0
			if items[0].QuoteFloatRate != nil {
				floatRate = *items[0].QuoteFloatRate
			}

			// 兼容百分比数值和纯小数表示
			rateVal := floatRate
			if rateVal >= 1.0 || rateVal <= -1.0 {
				rateVal = rateVal / 100.0
			}

			if rateVal < 0.05 {
				// 根据产品主分类（报价明细第一个.main_category，对应ProductCategoryName）
				category := ""
				if items[0].ProductCategoryName != nil {
					category = *items[0].ProductCategoryName
				}
				if category == "汽车涂层制动盘" {
					targetEmployeeID = 2
					targetEmployeeName = "李鹏涛"
				} else {
					targetEmployeeID = 81
					targetEmployeeName = "李永泉"
				}
			} else if rateVal >= 0.05 && rateVal < 0.10 {
				targetEmployeeID = 4
				targetEmployeeName = "廖语晴"
			} else if rateVal >= 0.10 {
				targetEmployeeID = 1
				targetEmployeeName = "肖鹏"
			}
		}

		// 用employee_id去general数据库employee查到姓名
		emp, err := qGen.Employee.WithContext(ctx).Where(qGen.Employee.ID.Eq(targetEmployeeID)).First()
		if err == nil && emp != nil && emp.Name != nil && *emp.Name != "" {
			targetEmployeeName = *emp.Name
		}

		// 6. 新建报价审批流节点记录：即quote_manage.quote_process_node
		node2Name := "报价审批"
		node2Status := "待审批"
		seqNum2 := int32(2)
		node2 := &quote_manage.QuoteProcessNode{
			ProcessID:           &process.ID,
			SeqNum:              &seqNum2,
			Name:                &node2Name,
			ApproveEmployeeID:   &targetEmployeeID,
			ApproveEmployeeName: &targetEmployeeName,
			Status:              &node2Status,
			CreatedAt:           &createdAtStr,
		}
		if err := qQuote.QuoteProcessNode.WithContext(ctx).Create(node2); err != nil {
			return err
		}

		// 7. 再回填刚刚新建报价单的一些字段（其实是指审批流记录 quote_process 的字段）
		process.PresentNodeID = &node2.ID
		process.PresentNodeName = &node2Name
		process.PresentApproverID = node2.ApproveEmployeeID
		process.PresentApproverName = node2.ApproveEmployeeName
		process.CreatedAt = &createdAtStr

		if _, err := qQuote.QuoteProcess.WithContext(ctx).Where(qQuote.QuoteProcess.ID.Eq(process.ID)).Updates(process); err != nil {
			return err
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
	return s.repo.SaveQuoteWithItems(ctx, quote, items, userID, userName)
}
