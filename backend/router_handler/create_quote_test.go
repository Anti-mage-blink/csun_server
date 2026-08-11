package router_handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"csun_server-backend/dao/model/quote_manage"
	"csun_server-backend/service_repository"

	"github.com/gin-gonic/gin"
)

type mockCreateQuoteServiceWrapper struct{}

func (m *mockCreateQuoteServiceWrapper) PrepareCreateQuote(ctx context.Context) (*service_repository.CreateQuoteData, error) {
	return &service_repository.CreateQuoteData{}, nil
}

func (m *mockCreateQuoteServiceWrapper) SubmitQuote(ctx context.Context, quote *quote_manage.Quote, items []*quote_manage.AQuoteItem, userID int32, userName string) error {
	if quote.AttachmentPathArray == nil || *quote.AttachmentPathArray != `["test_uploads/doc1.pdf","test_uploads/doc2.xlsx"]` {
		return fmt.Errorf("AttachmentPathArray conversion failed: %v", quote.AttachmentPathArray)
	}
	return nil
}

func TestSubmitQuoteHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("SubmitQuote with string array attachment_path_array", func(t *testing.T) {
		reqBody := gin.H{
			"quote": gin.H{
				"quote_code":            "BJ-20260811-001",
				"customer_name":         "测试客户",
				"pay_way":               "银行转账",
				"credit_period":         "30天",
				"attachment_path_array": []string{"test_uploads/doc1.pdf", "test_uploads/doc2.xlsx"},
			},
			"quote_items": []gin.H{
				{
					"product_name": "测试产品",
				},
			},
			"user": gin.H{
				"id":   1,
				"name": "张三",
			},
		}

		jsonBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/quote/submit", bytes.NewBuffer(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		h := NewCreateQuoteHandler(&mockCreateQuoteServiceWrapper{})

		r := gin.New()
		r.POST("/api/quote/submit", h.SubmitQuote)

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}
	})
}
