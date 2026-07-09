package service

import (
	"context"
	"errors"
	"testing"

	"csun_server-backend/dao/model/general"
)

// MockAuthRepository is a mock implementation of repository.AuthRepository
type MockAuthRepository struct {
	GetAccountByUsernameFn func(ctx context.Context, username string) (*general.Account, error)
	GetEmployeeByIDFn     func(ctx context.Context, employeeID int32) (*general.Employee, error)
}

func (m *MockAuthRepository) GetAccountByUsername(ctx context.Context, username string) (*general.Account, error) {
	return m.GetAccountByUsernameFn(ctx, username)
}

func (m *MockAuthRepository) GetEmployeeByID(ctx context.Context, employeeID int32) (*general.Employee, error) {
	return m.GetEmployeeByIDFn(ctx, employeeID)
}

func TestAuthService_Login(t *testing.T) {
	ctx := context.Background()

	t.Run("Login Successful", func(t *testing.T) {
		empID := int32(1)
		username := "admin"
		password := "password123"
		dept := "Engineering"
		name := "John Doe"

		mockRepo := &MockAuthRepository{
			GetAccountByUsernameFn: func(ctx context.Context, u string) (*general.Account, error) {
				return &general.Account{
					ID:         1,
					EmployeeID: &empID,
					Username:   &username,
					Password:   &password,
				}, nil
			},
			GetEmployeeByIDFn: func(ctx context.Context, id int32) (*general.Employee, error) {
				return &general.Employee{
					ID:         id,
					Name:       &name,
					Department: &dept,
				}, nil
			},
		}

		service := NewAuthService(mockRepo)
		employee, err := service.Login(ctx, "admin", "password123")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if employee == nil || *employee.Name != "John Doe" {
			t.Errorf("expected employee John Doe, got: %v", employee)
		}
	})

	t.Run("Password Incorrect", func(t *testing.T) {
		username := "admin"
		password := "correct_password"
		mockRepo := &MockAuthRepository{
			GetAccountByUsernameFn: func(ctx context.Context, u string) (*general.Account, error) {
				return &general.Account{
					ID:       1,
					Username: &username,
					Password: &password,
				}, nil
			},
		}

		service := NewAuthService(mockRepo)
		_, err := service.Login(ctx, "admin", "wrong_password")
		if !errors.Is(err, ErrPasswordIncorrect) {
			t.Errorf("expected ErrPasswordIncorrect, got: %v", err)
		}
	})

	t.Run("User Not Found", func(t *testing.T) {
		mockRepo := &MockAuthRepository{
			GetAccountByUsernameFn: func(ctx context.Context, u string) (*general.Account, error) {
				return nil, nil // Not found
			},
		}

		service := NewAuthService(mockRepo)
		_, err := service.Login(ctx, "nonexistent", "password")
		if !errors.Is(err, ErrUserNotFound) {
			t.Errorf("expected ErrUserNotFound, got: %v", err)
		}
	})

	t.Run("Employee ID is Nil", func(t *testing.T) {
		username := "admin"
		password := "password123"
		mockRepo := &MockAuthRepository{
			GetAccountByUsernameFn: func(ctx context.Context, u string) (*general.Account, error) {
				return &general.Account{
					ID:         1,
					EmployeeID: nil,
					Username:   &username,
					Password:   &password,
				}, nil
			},
		}

		service := NewAuthService(mockRepo)
		_, err := service.Login(ctx, "admin", "password123")
		if !errors.Is(err, ErrEmployeeNotFound) {
			t.Errorf("expected ErrEmployeeNotFound, got: %v", err)
		}
	})

	t.Run("Employee Not Found in DB", func(t *testing.T) {
		empID := int32(99)
		username := "admin"
		password := "password123"
		mockRepo := &MockAuthRepository{
			GetAccountByUsernameFn: func(ctx context.Context, u string) (*general.Account, error) {
				return &general.Account{
					ID:         1,
					EmployeeID: &empID,
					Username:   &username,
					Password:   &password,
				}, nil
			},
			GetEmployeeByIDFn: func(ctx context.Context, id int32) (*general.Employee, error) {
				return nil, nil // Not found
			},
		}

		service := NewAuthService(mockRepo)
		_, err := service.Login(ctx, "admin", "password123")
		if !errors.Is(err, ErrEmployeeNotFound) {
			t.Errorf("expected ErrEmployeeNotFound, got: %v", err)
		}
	})
}
