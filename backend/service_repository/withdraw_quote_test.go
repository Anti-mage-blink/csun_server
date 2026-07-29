package service_repository

import (
	"context"
	"errors"
	"testing"
)

type mockWithdrawQuoteRepository struct {
	WithdrawQuoteFn func(ctx context.Context, params *WithdrawQuoteRequestParams) error
}

func (m *mockWithdrawQuoteRepository) WithdrawQuote(ctx context.Context, params *WithdrawQuoteRequestParams) error {
	return m.WithdrawQuoteFn(ctx, params)
}

func TestWithdrawQuote_InvalidParams(t *testing.T) {
	svc := NewMockWithdrawQuoteService(&mockWithdrawQuoteRepository{})

	// 1. params 为 nil
	err := svc.WithdrawQuote(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil params, got nil")
	}

	// 2. process_id 不合法
	err = svc.WithdrawQuote(context.Background(), &WithdrawQuoteRequestParams{
		ProcessID: 0,
		UserID:    1,
		UserName:  "张三",
	})
	if err == nil {
		t.Fatal("expected error for process_id <= 0, got nil")
	}

	// 3. user_id 不合法
	err = svc.WithdrawQuote(context.Background(), &WithdrawQuoteRequestParams{
		ProcessID: 1,
		UserID:    0,
		UserName:  "张三",
	})
	if err == nil {
		t.Fatal("expected error for user_id <= 0, got nil")
	}
}

func TestWithdrawQuote_Success(t *testing.T) {
	called := false
	mockRepo := &mockWithdrawQuoteRepository{
		WithdrawQuoteFn: func(ctx context.Context, params *WithdrawQuoteRequestParams) error {
			called = true
			if params.ProcessID != 10 {
				return errors.New("expected process_id 10")
			}
			if params.UserID != 100 {
				return errors.New("expected user_id 100")
			}
			if params.UserName != "李四" {
				return errors.New("expected user_name 李四")
			}
			return nil
		},
	}

	svc := NewMockWithdrawQuoteService(mockRepo)
	err := svc.WithdrawQuote(context.Background(), &WithdrawQuoteRequestParams{
		ProcessID: 10,
		UserID:    100,
		UserName:  "李四",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatal("expected WithdrawQuote to be called")
	}
}

func TestWithdrawQuote_RepoError(t *testing.T) {
	mockRepo := &mockWithdrawQuoteRepository{
		WithdrawQuoteFn: func(ctx context.Context, params *WithdrawQuoteRequestParams) error {
			return errors.New("db delete error")
		},
	}

	svc := NewMockWithdrawQuoteService(mockRepo)
	err := svc.WithdrawQuote(context.Background(), &WithdrawQuoteRequestParams{
		ProcessID: 1,
		UserID:    1,
		UserName:  "测试",
	})

	if err == nil || err.Error() != "db delete error" {
		t.Fatalf("expected 'db delete error', got %v", err)
	}
}
