package router_handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"csun_server-backend/hr_sync/service_repository"

	"github.com/gin-gonic/gin"
)

func TestCallbackHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service_repository.NewCallbackService(nil)
	h := NewCallbackHandler(svc)
	r := gin.New()
	RegisterCallbackRoutes(r, h)

	t.Run("POST /api/hr_sync/callback with event_test payload", func(t *testing.T) {
		payload := `{"key":"event_test","data":[]}`
		req := httptest.NewRequest(http.MethodPost, "/api/hr_sync/callback", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var got service_repository.CallbackResult
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if got.ResultCode != "SUCCESS" || got.ResultMsg != "OK" {
			t.Errorf("expected {SUCCESS, OK}, got {%s, %s}", got.ResultCode, got.ResultMsg)
		}
	})

	t.Run("POST /api/hr_sync/callback missing key returns 400 FAIL", func(t *testing.T) {
		payload := `{"data":[]}`
		req := httptest.NewRequest(http.MethodPost, "/api/hr_sync/callback", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d, body: %s", w.Code, w.Body.String())
		}

		var got service_repository.CallbackResult
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if got.ResultCode != "FAIL" {
			t.Errorf("expected result_code FAIL, got %s", got.ResultCode)
		}
	})
}
