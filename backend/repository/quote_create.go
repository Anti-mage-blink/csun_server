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

// QuoteCreateRepository 新建报价单所需的数据库操作
type QuoteCreateRepository interface {
	GetMaxQuoteCodeByPrefix(ctx context.Context, prefix string) (*string, error)
	GetAllCustomers(ctx context.Context) ([]*general.Customer, error)
	GetAllProductCategories(ctx context.Context) ([]*product.ProductCategory, error)
	GetAllProductNames(ctx context.Context) ([]*product.ProductName, error)
	GetAllProductSpecs(ctx context.Context) ([]*product.ProductSpec, error)
	GetAllPriceCatalogs(ctx context.Context) ([]*product.PriceCatalog, error)
}

type quoteCreateRepository struct {
	db *gorm.DB
}

func NewQuoteCreateRepository(db *gorm.DB) QuoteCreateRepository {
	return &quoteCreateRepository{db: db}
}

// GetMaxQuoteCodeByPrefix 查询以 prefix 为前缀的最大 quote_code（按降序取首条）
func (r *quoteCreateRepository) GetMaxQuoteCodeByPrefix(ctx context.Context, prefix string) (*string, error) {
	q := quote_query.Use(r.db)
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

func (r *quoteCreateRepository) GetAllCustomers(ctx context.Context) ([]*general.Customer, error) {
	q := general_query.Use(r.db)
	return q.Customer.WithContext(ctx).Find()
}

func (r *quoteCreateRepository) GetAllProductCategories(ctx context.Context) ([]*product.ProductCategory, error) {
	q := product_query.Use(r.db)
	return q.ProductCategory.WithContext(ctx).Find()
}

func (r *quoteCreateRepository) GetAllProductNames(ctx context.Context) ([]*product.ProductName, error) {
	q := product_query.Use(r.db)
	return q.ProductName.WithContext(ctx).Find()
}

func (r *quoteCreateRepository) GetAllProductSpecs(ctx context.Context) ([]*product.ProductSpec, error) {
	q := product_query.Use(r.db)
	return q.ProductSpec.WithContext(ctx).Find()
}

func (r *quoteCreateRepository) GetAllPriceCatalogs(ctx context.Context) ([]*product.PriceCatalog, error) {
	q := product_query.Use(r.db)
	return q.PriceCatalog.WithContext(ctx).Find()
}
