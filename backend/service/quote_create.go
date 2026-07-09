package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"csun_server-backend/dao/model/general"
	"csun_server-backend/dao/model/product"
	"csun_server-backend/repository"
)

// QuoteCreateData 新建报价单返回给前端的全量数据
type QuoteCreateData struct {
	QuoteCode         string                     `json:"quote_code"`
	Customers         []*general.Customer        `json:"customers"`
	ProductCategories []*product.ProductCategory `json:"product_categories"`
	ProductNames      []*product.ProductName     `json:"product_names"`
	ProductSpecs      []*product.ProductSpec     `json:"product_specs"`
	PriceCatalogs     []*product.PriceCatalog    `json:"price_catalogs"`
}

type QuoteCreateService interface {
	PrepareQuoteCreate(ctx context.Context) (*QuoteCreateData, error)
}

type quoteCreateService struct {
	quoteCreateRepo repository.QuoteCreateRepository
}

func NewQuoteCreateService(quoteCreateRepo repository.QuoteCreateRepository) QuoteCreateService {
	return &quoteCreateService{quoteCreateRepo: quoteCreateRepo}
}

func (s *quoteCreateService) PrepareQuoteCreate(ctx context.Context) (*QuoteCreateData, error) {
	// 1. 生成报价单编号：前缀 "BJ-{今日年月日}-"
	today := time.Now().Format("20060102")
	prefix := fmt.Sprintf("BJ-%s-", today)

	maxCode, err := s.quoteCreateRepo.GetMaxQuoteCodeByPrefix(ctx, prefix)
	if err != nil {
		return nil, err
	}
	quoteCode := generateQuoteCode(prefix, maxCode)

	// 2. 获取全量数据表，方便前端使用
	customers, err := s.quoteCreateRepo.GetAllCustomers(ctx)
	if err != nil {
		return nil, err
	}
	productCategories, err := s.quoteCreateRepo.GetAllProductCategories(ctx)
	if err != nil {
		return nil, err
	}
	productNames, err := s.quoteCreateRepo.GetAllProductNames(ctx)
	if err != nil {
		return nil, err
	}
	productSpecs, err := s.quoteCreateRepo.GetAllProductSpecs(ctx)
	if err != nil {
		return nil, err
	}
	priceCatalogs, err := s.quoteCreateRepo.GetAllPriceCatalogs(ctx)
	if err != nil {
		return nil, err
	}

	return &QuoteCreateData{
		QuoteCode:         quoteCode,
		Customers:         customers,
		ProductCategories: productCategories,
		ProductNames:      productNames,
		ProductSpecs:      productSpecs,
		PriceCatalogs:     priceCatalogs,
	}, nil
}

// generateQuoteCode 根据前缀和当前最大编号生成新的报价单编号
// 若无历史记录，返回 "{prefix}001"；否则将最后一段序号 +1（3 位补零）
func generateQuoteCode(prefix string, maxCode *string) string {
	const defaultSeq = "001"
	if maxCode == nil {
		return prefix + defaultSeq
	}
	// maxCode 格式: "BJ-20260709-001"，截取前缀之后的部分即为序号
	seqStr := (*maxCode)[len(prefix):]
	seq, err := strconv.Atoi(seqStr)
	if err != nil {
		return prefix + defaultSeq
	}
	return fmt.Sprintf("%s%03d", prefix, seq+1)
}
