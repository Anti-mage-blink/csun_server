package service_repository

import (
	"context"
	"errors"
	"testing"

	"csun_server-backend/dao/model/general"
	"csun_server-backend/dao/model/quote_manage"
)

// MockCreateQuoteRepository is a mock implementation of createQuoteRepository
type MockCreateQuoteRepository struct {
	GetMaxQuoteCodeByPrefixFn func(ctx context.Context, prefix string) (*string, error)
	GetAllCustomersFn         func(ctx context.Context) ([]*general.Customer, error)
	GetAllProductSpecsFn      func(ctx context.Context) ([]*quote_manage.ProductSpec, error)
}

func (m *MockCreateQuoteRepository) GetMaxQuoteCodeByPrefix(ctx context.Context, prefix string) (*string, error) {
	return m.GetMaxQuoteCodeByPrefixFn(ctx, prefix)
}

func (m *MockCreateQuoteRepository) GetAllCustomers(ctx context.Context) ([]*general.Customer, error) {
	return m.GetAllCustomersFn(ctx)
}

func (m *MockCreateQuoteRepository) GetAllProductSpecs(ctx context.Context) ([]*quote_manage.ProductSpec, error) {
	return m.GetAllProductSpecsFn(ctx)
}

func (m *MockCreateQuoteRepository) SaveQuoteWithItems(ctx context.Context, quote *quote_manage.Quote, items []*quote_manage.AQuoteItem, userID int32, userName string) error {
	return nil
}

// newDefaultMockRepo 返回一个所有查询都返回空切片成功的 mock，便于聚焦测试核心逻辑
func newDefaultMockRepo() *MockCreateQuoteRepository {
	return &MockCreateQuoteRepository{
		GetMaxQuoteCodeByPrefixFn: func(ctx context.Context, prefix string) (*string, error) {
			return nil, nil
		},
		GetAllCustomersFn:    func(ctx context.Context) ([]*general.Customer, error) { return nil, nil },
		GetAllProductSpecsFn: func(ctx context.Context) ([]*quote_manage.ProductSpec, error) { return nil, nil },
	}
}

func TestCreateQuoteService_PrepareCreateQuote(t *testing.T) {
	ctx := context.Background()

	t.Run("No existing quote code - returns 001", func(t *testing.T) {
		mockRepo := newDefaultMockRepo()
		svc := NewMockCreateQuoteService(mockRepo)

		data, err := svc.PrepareCreateQuote(ctx)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if data.Quote == nil || data.Quote.QuoteCode == nil {
			t.Fatalf("expected Quote and QuoteCode to be populated")
		}
		quoteCode := *data.Quote.QuoteCode
		// 校验编号以 "BJ-" 开头并以 "-001" 结尾
		if len(quoteCode) < len("BJ-20060102-001") {
			t.Errorf("quote code too short: %s", quoteCode)
		}
		if quoteCode[len(quoteCode)-4:] != "-001" {
			t.Errorf("expected quote code ending with -001, got: %s", quoteCode)
		}
	})

	t.Run("Existing quote code - increments sequence", func(t *testing.T) {
		existing := "BJ-20260709-005"
		mockRepo := newDefaultMockRepo()
		mockRepo.GetMaxQuoteCodeByPrefixFn = func(ctx context.Context, prefix string) (*string, error) {
			return &existing, nil
		}
		svc := NewMockCreateQuoteService(mockRepo)

		data, err := svc.PrepareCreateQuote(ctx)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if data.Quote == nil || data.Quote.QuoteCode == nil {
			t.Fatalf("expected Quote and QuoteCode to be populated")
		}
		got := *data.Quote.QuoteCode
		if got[len(got)-4:] != "-006" {
			t.Errorf("expected sequence -006, got: %s", got)
		}
	})

	t.Run("Existing quote code with sequence 999 - rolls to 1000", func(t *testing.T) {
		existing := "BJ-20260709-999"
		mockRepo := newDefaultMockRepo()
		mockRepo.GetMaxQuoteCodeByPrefixFn = func(ctx context.Context, prefix string) (*string, error) {
			return &existing, nil
		}
		svc := NewMockCreateQuoteService(mockRepo)

		data, err := svc.PrepareCreateQuote(ctx)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if data.Quote == nil || data.Quote.QuoteCode == nil {
			t.Fatalf("expected Quote and QuoteCode to be populated")
		}
		got := *data.Quote.QuoteCode
		if got[len(got)-5:] != "-1000" {
			t.Errorf("expected sequence -1000, got: %s", got)
		}
	})

	t.Run("Returns all base data tables", func(t *testing.T) {
		customerName := "Acme"
		spec := "10x2"

		mockRepo := newDefaultMockRepo()
		mockRepo.GetAllCustomersFn = func(ctx context.Context) ([]*general.Customer, error) {
			return []*general.Customer{{ID: 1, CompanyName: &customerName}}, nil
		}
		mockRepo.GetAllProductSpecsFn = func(ctx context.Context) ([]*quote_manage.ProductSpec, error) {
			return []*quote_manage.ProductSpec{{ID: 1, Spec: &spec}}, nil
		}
		svc := NewMockCreateQuoteService(mockRepo)

		data, err := svc.PrepareCreateQuote(ctx)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(data.Customers) != 1 || *data.Customers[0].CompanyName != "Acme" {
			t.Errorf("unexpected customers: %v", data.Customers)
		}
		if len(data.ProductSpecs) != 1 || *data.ProductSpecs[0].Spec != "10x2" {
			t.Errorf("unexpected product specs: %v", data.ProductSpecs)
		}
	})

	t.Run("Repository error propagates", func(t *testing.T) {
		repoErr := errors.New("db error")
		mockRepo := newDefaultMockRepo()
		mockRepo.GetMaxQuoteCodeByPrefixFn = func(ctx context.Context, prefix string) (*string, error) {
			return nil, repoErr
		}
		svc := NewMockCreateQuoteService(mockRepo)

		_, err := svc.PrepareCreateQuote(ctx)
		if !errors.Is(err, repoErr) {
			t.Errorf("expected repo error to propagate, got: %v", err)
		}
	})
}

func TestCreateQuoteService_SubmitQuote(t *testing.T) {
	ctx := context.Background()

	t.Run("SubmitQuote nil quote returns error", func(t *testing.T) {
		mockRepo := newDefaultMockRepo()
		svc := NewMockCreateQuoteService(mockRepo)

		err := svc.SubmitQuote(ctx, nil, []*quote_manage.AQuoteItem{{}}, 1, "test")
		if err == nil {
			t.Error("expected error for nil quote")
		}
	})

	t.Run("SubmitQuote empty items returns error", func(t *testing.T) {
		mockRepo := newDefaultMockRepo()
		svc := NewMockCreateQuoteService(mockRepo)

		err := svc.SubmitQuote(ctx, &quote_manage.Quote{}, []*quote_manage.AQuoteItem{}, 1, "test")
		if err == nil {
			t.Error("expected error for empty items")
		}
	})

	t.Run("SubmitQuote missing pay_way or credit_period returns error", func(t *testing.T) {
		mockRepo := newDefaultMockRepo()
		svc := NewMockCreateQuoteService(mockRepo)

		payWay := "银行转账"
		creditPeriod := "30天"

		// 缺少 pay_way
		err := svc.SubmitQuote(ctx, &quote_manage.Quote{CreditPeriod: &creditPeriod}, []*quote_manage.AQuoteItem{{}}, 1, "test")
		if err == nil {
			t.Error("expected error for missing pay_way")
		}

		// 缺少 credit_period
		err = svc.SubmitQuote(ctx, &quote_manage.Quote{PayWay: &payWay}, []*quote_manage.AQuoteItem{{}}, 1, "test")
		if err == nil {
			t.Error("expected error for missing credit_period")
		}
	})

	t.Run("SubmitQuote success with all 4 new fields", func(t *testing.T) {
		mockRepo := newDefaultMockRepo()
		svc := NewMockCreateQuoteService(mockRepo)

		compName := "测试公司"
		contName := "张三"
		contTitle := "经理"
		payWay := "银行转账"
		creditPeriod := "30天"
		remarks := "加急订单"
		attachmentPathArray := `["/upload/doc1.pdf"]`

		q := &quote_manage.Quote{
			CustomerName:        &compName,
			ContactName:         &contName,
			ContactTitle:        &contTitle,
			PayWay:              &payWay,
			CreditPeriod:        &creditPeriod,
			Remarks:             &remarks,
			AttachmentPathArray: &attachmentPathArray,
		}
		items := []*quote_manage.AQuoteItem{{}}

		err := svc.SubmitQuote(ctx, q, items, 1, "测试人员")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}
