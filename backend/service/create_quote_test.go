package service

import (
	"context"
	"errors"
	"testing"

	"csun_server-backend/dao/model/general"
	"csun_server-backend/dao/model/product"
)

// MockCreateQuoteRepository is a mock implementation of repository.CreateQuoteRepository
type MockCreateQuoteRepository struct {
	GetMaxQuoteCodeByPrefixFn func(ctx context.Context, prefix string) (*string, error)
	GetAllCustomersFn         func(ctx context.Context) ([]*general.Customer, error)
	GetAllProductCategoriesFn func(ctx context.Context) ([]*product.ProductCategory, error)
	GetAllProductNamesFn      func(ctx context.Context) ([]*product.ProductName, error)
	GetAllProductSpecsFn      func(ctx context.Context) ([]*product.ProductSpec, error)
	GetAllPriceCatalogsFn     func(ctx context.Context) ([]*product.PriceCatalog, error)
}

func (m *MockCreateQuoteRepository) GetMaxQuoteCodeByPrefix(ctx context.Context, prefix string) (*string, error) {
	return m.GetMaxQuoteCodeByPrefixFn(ctx, prefix)
}

func (m *MockCreateQuoteRepository) GetAllCustomers(ctx context.Context) ([]*general.Customer, error) {
	return m.GetAllCustomersFn(ctx)
}

func (m *MockCreateQuoteRepository) GetAllProductCategories(ctx context.Context) ([]*product.ProductCategory, error) {
	return m.GetAllProductCategoriesFn(ctx)
}

func (m *MockCreateQuoteRepository) GetAllProductNames(ctx context.Context) ([]*product.ProductName, error) {
	return m.GetAllProductNamesFn(ctx)
}

func (m *MockCreateQuoteRepository) GetAllProductSpecs(ctx context.Context) ([]*product.ProductSpec, error) {
	return m.GetAllProductSpecsFn(ctx)
}

func (m *MockCreateQuoteRepository) GetAllPriceCatalogs(ctx context.Context) ([]*product.PriceCatalog, error) {
	return m.GetAllPriceCatalogsFn(ctx)
}

// newDefaultMockRepo 返回一个所有查询都返回空切片成功的 mock，便于聚焦测试核心逻辑
func newDefaultMockRepo() *MockCreateQuoteRepository {
	return &MockCreateQuoteRepository{
		GetMaxQuoteCodeByPrefixFn: func(ctx context.Context, prefix string) (*string, error) {
			return nil, nil
		},
		GetAllCustomersFn:         func(ctx context.Context) ([]*general.Customer, error) { return nil, nil },
		GetAllProductCategoriesFn: func(ctx context.Context) ([]*product.ProductCategory, error) { return nil, nil },
		GetAllProductNamesFn:      func(ctx context.Context) ([]*product.ProductName, error) { return nil, nil },
		GetAllProductSpecsFn:      func(ctx context.Context) ([]*product.ProductSpec, error) { return nil, nil },
		GetAllPriceCatalogsFn:     func(ctx context.Context) ([]*product.PriceCatalog, error) { return nil, nil },
	}
}

func TestCreateQuoteService_PrepareCreateQuote(t *testing.T) {
	ctx := context.Background()

	t.Run("No existing quote code - returns 001", func(t *testing.T) {
		mockRepo := newDefaultMockRepo()
		svc := NewCreateQuoteService(mockRepo)

		data, err := svc.PrepareCreateQuote(ctx)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		// 校验编号以 "BJ-" 开头并以 "-001" 结尾
		if len(data.QuoteCode) < len("BJ-20060102-001") {
			t.Errorf("quote code too short: %s", data.QuoteCode)
		}
		if data.QuoteCode[len(data.QuoteCode)-4:] != "-001" {
			t.Errorf("expected quote code ending with -001, got: %s", data.QuoteCode)
		}
	})

	t.Run("Existing quote code - increments sequence", func(t *testing.T) {
		existing := "BJ-20260709-005"
		mockRepo := newDefaultMockRepo()
		mockRepo.GetMaxQuoteCodeByPrefixFn = func(ctx context.Context, prefix string) (*string, error) {
			return &existing, nil
		}
		svc := NewCreateQuoteService(mockRepo)

		data, err := svc.PrepareCreateQuote(ctx)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		// 注意：今日日期由 time.Now() 决定，故只校验序号递增部分
		got := data.QuoteCode
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
		svc := NewCreateQuoteService(mockRepo)

		data, err := svc.PrepareCreateQuote(ctx)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		got := data.QuoteCode
		if got[len(got)-5:] != "-1000" {
			t.Errorf("expected sequence -1000, got: %s", got)
		}
	})

	t.Run("Returns all base data tables", func(t *testing.T) {
		customerName := "Acme"
		categoryName := "Rubber"
		productName := "O-Ring"
		spec := "10x2"
		price := 1.23

		mockRepo := newDefaultMockRepo()
		mockRepo.GetAllCustomersFn = func(ctx context.Context) ([]*general.Customer, error) {
			return []*general.Customer{{ID: 1, CompanyName: &customerName}}, nil
		}
		mockRepo.GetAllProductCategoriesFn = func(ctx context.Context) ([]*product.ProductCategory, error) {
			return []*product.ProductCategory{{ID: 1, Name: &categoryName}}, nil
		}
		mockRepo.GetAllProductNamesFn = func(ctx context.Context) ([]*product.ProductName, error) {
			return []*product.ProductName{{ID: 1, Name: &productName}}, nil
		}
		mockRepo.GetAllProductSpecsFn = func(ctx context.Context) ([]*product.ProductSpec, error) {
			return []*product.ProductSpec{{ID: 1, Spec: &spec}}, nil
		}
		mockRepo.GetAllPriceCatalogsFn = func(ctx context.Context) ([]*product.PriceCatalog, error) {
			return []*product.PriceCatalog{{ID: 1, BigBatchBasePrice: &price}}, nil
		}
		svc := NewCreateQuoteService(mockRepo)

		data, err := svc.PrepareCreateQuote(ctx)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if len(data.Customers) != 1 || *data.Customers[0].CompanyName != "Acme" {
			t.Errorf("unexpected customers: %v", data.Customers)
		}
		if len(data.ProductCategories) != 1 || *data.ProductCategories[0].Name != "Rubber" {
			t.Errorf("unexpected product categories: %v", data.ProductCategories)
		}
		if len(data.ProductNames) != 1 || *data.ProductNames[0].Name != "O-Ring" {
			t.Errorf("unexpected product names: %v", data.ProductNames)
		}
		if len(data.ProductSpecs) != 1 || *data.ProductSpecs[0].Spec != "10x2" {
			t.Errorf("unexpected product specs: %v", data.ProductSpecs)
		}
		if len(data.PriceCatalogs) != 1 || *data.PriceCatalogs[0].BigBatchBasePrice != 1.23 {
			t.Errorf("unexpected price catalogs: %v", data.PriceCatalogs)
		}
	})

	t.Run("Repository error propagates", func(t *testing.T) {
		repoErr := errors.New("db error")
		mockRepo := newDefaultMockRepo()
		mockRepo.GetMaxQuoteCodeByPrefixFn = func(ctx context.Context, prefix string) (*string, error) {
			return nil, repoErr
		}
		svc := NewCreateQuoteService(mockRepo)

		_, err := svc.PrepareCreateQuote(ctx)
		if !errors.Is(err, repoErr) {
			t.Errorf("expected repo error to propagate, got: %v", err)
		}
	})
}

func TestGenerateQuoteCode(t *testing.T) {
	prefix := "BJ-20260709-"

	t.Run("nil max code", func(t *testing.T) {
		if got := generateQuoteCode(prefix, nil); got != "BJ-20260709-001" {
			t.Errorf("expected BJ-20260709-001, got: %s", got)
		}
	})

	t.Run("increment normal", func(t *testing.T) {
		existing := "BJ-20260709-001"
		if got := generateQuoteCode(prefix, &existing); got != "BJ-20260709-002" {
			t.Errorf("expected BJ-20260709-002, got: %s", got)
		}
	})

	t.Run("increment with carry", func(t *testing.T) {
		existing := "BJ-20260709-009"
		if got := generateQuoteCode(prefix, &existing); got != "BJ-20260709-010" {
			t.Errorf("expected BJ-20260709-010, got: %s", got)
		}
	})

	t.Run("non-numeric sequence falls back to 001", func(t *testing.T) {
		existing := "BJ-20260709-abc"
		if got := generateQuoteCode(prefix, &existing); got != "BJ-20260709-001" {
			t.Errorf("expected BJ-20260709-001, got: %s", got)
		}
	})
}
