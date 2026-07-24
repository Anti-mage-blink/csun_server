package service_repository

import (
	"context"
	"errors"
	"testing"
)

type mockApproveHandleRepository struct {
	UpdateNodeAndProcessStatusFn func(ctx context.Context, params *ApproveHandleRequestParams) error
}

func (m *mockApproveHandleRepository) UpdateNodeAndProcessStatus(ctx context.Context, params *ApproveHandleRequestParams) error {
	return m.UpdateNodeAndProcessStatusFn(ctx, params)
}

func TestApproveHandle_InvalidParams(t *testing.T) {
	svc := NewMockApproveHandleService(&mockApproveHandleRepository{})

	// 1. params 为 nil
	err := svc.ApproveHandle(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil params, got nil")
	}

	// 2. action 不合法
	err = svc.ApproveHandle(context.Background(), &ApproveHandleRequestParams{
		Action:    "invalid",
		NodeID:    1,
		ProcessID: 1,
	})
	if err == nil {
		t.Fatal("expected error for invalid action, got nil")
	}

	// 3. node_id 不合法
	err = svc.ApproveHandle(context.Background(), &ApproveHandleRequestParams{
		Action:    "approve",
		NodeID:    0,
		ProcessID: 1,
	})
	if err == nil {
		t.Fatal("expected error for node_id = 0, got nil")
	}

	// 4. process_id 不合法
	err = svc.ApproveHandle(context.Background(), &ApproveHandleRequestParams{
		Action:    "reject",
		NodeID:    1,
		ProcessID: -1,
	})
	if err == nil {
		t.Fatal("expected error for process_id < 0, got nil")
	}
}

func TestApproveHandle_Success(t *testing.T) {
	called := false
	mockRepo := &mockApproveHandleRepository{
		UpdateNodeAndProcessStatusFn: func(ctx context.Context, params *ApproveHandleRequestParams) error {
			called = true
			if params.Action != "approve" {
				return errors.New("expected approve")
			}
			if params.NodeID != 10 {
				return errors.New("expected node_id 10")
			}
			if params.ProcessID != 20 {
				return errors.New("expected process_id 20")
			}
			if params.Comment != "ok" {
				return errors.New("expected comment 'ok'")
			}
			return nil
		},
	}

	svc := NewMockApproveHandleService(mockRepo)
	err := svc.ApproveHandle(context.Background(), &ApproveHandleRequestParams{
		Action:    "approve",
		NodeID:    10,
		ProcessID: 20,
		Comment:   "ok",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatal("expected UpdateNodeAndProcessStatus to be called")
	}
}

func TestApproveHandle_RepoError(t *testing.T) {
	mockRepo := &mockApproveHandleRepository{
		UpdateNodeAndProcessStatusFn: func(ctx context.Context, params *ApproveHandleRequestParams) error {
			return errors.New("db write error")
		},
	}

	svc := NewMockApproveHandleService(mockRepo)
	err := svc.ApproveHandle(context.Background(), &ApproveHandleRequestParams{
		Action:    "reject",
		NodeID:    10,
		ProcessID: 20,
		Comment:   "not ok",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "db write error" {
		t.Errorf("expected 'db write error', got '%v'", err)
	}
}
