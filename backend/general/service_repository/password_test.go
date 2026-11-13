package service_repository

import (
	"context"
	"errors"
	"testing"

	"csun_server-backend/dao/model/general"
)

type MockPasswordRepository struct {
	GetAccountByUsernameFn func(ctx context.Context, username string) (*general.Account, error)
	UpdatePasswordFn       func(ctx context.Context, account *general.Account, newPassword string) error
}

func (m *MockPasswordRepository) GetAccountByUsername(ctx context.Context, username string) (*general.Account, error) {
	return m.GetAccountByUsernameFn(ctx, username)
}

func (m *MockPasswordRepository) UpdatePassword(ctx context.Context, account *general.Account, newPassword string) error {
	return m.UpdatePasswordFn(ctx, account, newPassword)
}

func TestPasswordService_ModifyPassword(t *testing.T) {
	ctx := context.Background()

	t.Run("Modify Password Success (Passwords are identical)", func(t *testing.T) {
		username := "admin"
		password := "password123"
		account := &general.Account{
			ID:       1,
			Username: &username,
			Password: &password,
		}

		var updatedPassword string
		mockRepo := &MockPasswordRepository{
			GetAccountByUsernameFn: func(ctx context.Context, u string) (*general.Account, error) {
				return account, nil
			},
			UpdatePasswordFn: func(ctx context.Context, acc *general.Account, newPass string) error {
				updatedPassword = newPass
				return nil
			},
		}

		service := NewMockPasswordService(mockRepo)
		err := service.ModifyPassword(ctx, "admin", "password123")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if updatedPassword != "password123" {
			t.Errorf("expected updated password to be password123, got: %s", updatedPassword)
		}
	})

	t.Run("Passwords are different", func(t *testing.T) {
		username := "admin"
		password := "old_password"
		account := &general.Account{
			ID:       1,
			Username: &username,
			Password: &password,
		}

		mockRepo := &MockPasswordRepository{
			GetAccountByUsernameFn: func(ctx context.Context, u string) (*general.Account, error) {
				return account, nil
			},
		}

		service := NewMockPasswordService(mockRepo)
		err := service.ModifyPassword(ctx, "admin", "new_password")
		if !errors.Is(err, ErrPasswordDifferent) {
			t.Errorf("expected ErrPasswordDifferent, got: %v", err)
		}
	})

	t.Run("User Not Found", func(t *testing.T) {
		mockRepo := &MockPasswordRepository{
			GetAccountByUsernameFn: func(ctx context.Context, u string) (*general.Account, error) {
				return nil, nil
			},
		}

		service := NewMockPasswordService(mockRepo)
		err := service.ModifyPassword(ctx, "nonexistent", "password123")
		if !errors.Is(err, ErrUserNotFound) {
			t.Errorf("expected ErrUserNotFound, got: %v", err)
		}
	})
}
