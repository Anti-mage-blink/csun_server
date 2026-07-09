package service

import (
	"context"
	"errors"
	"csun_server-backend/repository"
)

var (
	ErrPasswordDifferent = errors.New("请设置新密码")
)

type PasswordService interface {
	ModifyPassword(ctx context.Context, username, password string) error
}

type passwordService struct {
	passwordRepo repository.PasswordRepository
}

func NewPasswordService(passwordRepo repository.PasswordRepository) PasswordService {
	return &passwordService{passwordRepo: passwordRepo}
}

func (s *passwordService) ModifyPassword(ctx context.Context, username, password string) error {
	// 1. 在 account 表中查找是否有 username 字段值为传入用户名的记录
	account, err := s.passwordRepo.GetAccountByUsername(ctx, username)
	if err != nil {
		return err
	}
	if account == nil {
		return ErrUserNotFound // 复用已有的 "该用户不存在"
	}

	// 2. 若密码相同
	if account.Password != nil && *account.Password == password {
		// 则修改该记录的 password 字段值为传入的密码
		err = s.passwordRepo.UpdatePassword(ctx, account, password)
		if err != nil {
			return err
		}
		return nil // 修改成功
	}

	// 3. 若密码不同
	return ErrPasswordDifferent
}
