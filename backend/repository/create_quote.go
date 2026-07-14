package repository

import (
	"context"
	"errors"

	"csun_server-backend/dao/model/general"
	"csun_server-backend/dao/model/product"
	general_query "csun_server-backend/dao/query/general"
	product_query "csun_server-backend/dao/query/product"
	quote_query "csun_server-backend/dao/query/quote_manage"

	"gorm.io/gorm"
)

// CreateQuoteRepository 新建报价单所需的数据库操作
type CreateQuoteRepository interface {
	GetMaxQuoteCodeByPrefix(ctx context.Context, prefix string) (*string, error)
	GetAllCustomers(ctx context.Context) ([]*general.Customer, error)
	GetAllProductCategories(ctx context.Context) ([]*product.ProductCategory, error)
	GetAllProductNames(ctx context.Context) ([]*product.ProductName, error)
	GetAllProductSpecs(ctx context.Context) ([]*product.ProductSpec, error)
	GetAllPriceCatalogs(ctx context.Context) ([]*product.PriceCatalog, error)
}

type createQuoteRepository struct {
	db *DBEngine
}

func NewCreateQuoteRepository(db *DBEngine) CreateQuoteRepository {
	return &createQuoteRepository{db: db}
}

// GetMaxQuoteCodeByPrefix 查询以 prefix 为前缀的最大 quote_code（按降序取首条）
func (r *createQuoteRepository) GetMaxQuoteCodeByPrefix(ctx context.Context, prefix string) (*string, error) {
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

func (r *createQuoteRepository) GetAllCustomers(ctx context.Context) ([]*general.Customer, error) {
	q := general_query.Use(r.db.General)
	return q.Customer.WithContext(ctx).Find()
}

func (r *createQuoteRepository) GetAllProductCategories(ctx context.Context) ([]*product.ProductCategory, error) {
	q := product_query.Use(r.db.Product)
	return q.ProductCategory.WithContext(ctx).Find()
}

func (r *createQuoteRepository) GetAllProductNames(ctx context.Context) ([]*product.ProductName, error) {
	q := product_query.Use(r.db.Product)
	return q.ProductName.WithContext(ctx).Find()
}

func (r *createQuoteRepository) GetAllProductSpecs(ctx context.Context) ([]*product.ProductSpec, error) {
	q := product_query.Use(r.db.Product)
	return q.ProductSpec.WithContext(ctx).Find()
}

func (r *createQuoteRepository) GetAllPriceCatalogs(ctx context.Context) ([]*product.PriceCatalog, error) {
	q := product_query.Use(r.db.Product)
	return q.PriceCatalog.WithContext(ctx).Find()
}
