package service_repository

import (
	"context"
	"encoding/json"
	"testing"
)

type mockCallbackRepository struct{}

func TestCallbackService_HandleCallback(t *testing.T) {
	ctx := context.Background()
	svc := NewMockCallbackService(&mockCallbackRepository{})

	t.Run("event_test returns SUCCESS and OK", func(t *testing.T) {
		res, err := svc.HandleCallback(ctx, &CallbackParams{
			Key:  "event_test",
			Data: json.RawMessage(`[]`),
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if res == nil {
			t.Fatal("expected non-nil result")
		}
		if res.ResultCode != "SUCCESS" || res.ResultMsg != "OK" {
			t.Errorf("expected {SUCCESS, OK}, got {%s, %s}", res.ResultCode, res.ResultMsg)
		}
	})

	t.Run("other event key returns SUCCESS and OK", func(t *testing.T) {
		res, err := svc.HandleCallback(ctx, &CallbackParams{
			Key:  "employee_add",
			Data: json.RawMessage(`["emp-id-1","emp-id-2"]`),
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if res.ResultCode != "SUCCESS" || res.ResultMsg != "OK" {
			t.Errorf("expected {SUCCESS, OK}, got {%s, %s}", res.ResultCode, res.ResultMsg)
		}
	})

	t.Run("nil params or empty key returns error", func(t *testing.T) {
		if _, err := svc.HandleCallback(ctx, nil); err == nil {
			t.Error("expected error for nil params, got nil")
		}
		if _, err := svc.HandleCallback(ctx, &CallbackParams{Key: ""}); err == nil {
			t.Error("expected error for empty key, got nil")
		}
	})
}
