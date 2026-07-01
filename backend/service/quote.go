package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"csun_server-backend/dao/query/quote_manage"

	"gorm.io/gorm"
)

// GetNewQuoteCode 生成新的报价单编号
// 逻辑：
// 1. 查询 quote_manage 数据库的 quote 数据表的 “quote_code” 字段（以 “BJ-${今日年月日（如20260630）}” 为前缀）
// 2. 返回：
//    1. 若查到，将查到的字符串（如"BJ-${今日年月日}-001"）最后一段 + 1（格式化为三位数字），返回
//    2. 若没查到，返回 "BJ-${今日年月日}-001" 字符串
func GetNewQuoteCode(ctx context.Context, db *gorm.DB) (string, error) {
	// 获取当前系统时间的今日年月日格式，例如 "20260630"
	todayStr := time.Now().Format("20060102")
	prefix := "BJ-" + todayStr

	// 初始化 quote_manage 查询实例
	q := quote_manage.Use(db).Quote

	// 查找以 prefix 开头的最大 quote_code，按 quote_code 降序排序
	lastQuote, err := q.WithContext(ctx).
		Where(q.QuoteCode.Like(prefix + "%")).
		Order(q.QuoteCode.Desc()).
		First()

	if err != nil {
		// 如果是未找到记录，正常返回 -001
		if err == gorm.ErrRecordNotFound {
			return prefix + "-001", nil
		}
		return "", fmt.Errorf("查询数据库已有报价单号失败: %w", err)
	}

	// 检查查询到的报价单对象及报价单号指针
	if lastQuote == nil || lastQuote.QuoteCode == nil || *lastQuote.QuoteCode == "" {
		return prefix + "-001", nil
	}

	code := *lastQuote.QuoteCode
	parts := strings.Split(code, "-")
	if len(parts) < 3 {
		// 如果格式无法正确分割为至少3段（BJ - YYYYMMDD - SEQ），退避返回 001
		return prefix + "-001", nil
	}

	// 取最后一段序列号
	lastSeqStr := parts[len(parts)-1]
	seq, err := strconv.Atoi(lastSeqStr)
	if err != nil {
		// 如果序列号部分非纯数字，退避返回 001
		return prefix + "-001", nil
	}

	// 自增 1
	nextSeq := seq + 1
	return fmt.Sprintf("%s-%03d", prefix, nextSeq), nil
}
