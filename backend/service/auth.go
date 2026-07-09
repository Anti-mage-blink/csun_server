package service

import (
	"context"
	"errors"
	"csun_server-backend/dao/model/general"
	"csun_server-backend/repository"
)

var (
	ErrPasswordIncorrect = errors.New("密码错误")
	ErrUserNotFound      = errors.New("该用户不存在")
	ErrEmployeeNotFound  = errors.New("关联的员工不存在")
)

type AuthService interface {
	Login(ctx context.Context, username, password string) (*general.Employee, error)
}

type authService struct {
	authRepo repository.AuthRepository
}

func NewAuthService(authRepo repository.AuthRepository) AuthService {
	return &authService{authRepo: authRepo}
}

func (s *authService) Login(ctx context.Context, username, password string) (*general.Employee, error) {
	// 1. 查找是否有该 username
	account, err := s.authRepo.GetAccountByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, ErrUserNotFound
	}

	// 2. 校验密码是否正确
	if account.Password == nil || *account.Password != password {
		return nil, ErrPasswordIncorrect
	}

	// 3. 根据 employee_id 查找关联员工
	if account.EmployeeID == nil {
		return nil, ErrEmployeeNotFound
	}

	employee, err := s.authRepo.GetEmployeeByID(ctx, *account.EmployeeID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	return employee, nil
}
