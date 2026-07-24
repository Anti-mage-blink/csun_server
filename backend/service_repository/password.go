package service_repository

import (
	"context"
	"csun_server-backend/dao/model/general"
	general_query "csun_server-backend/dao/query/general"
	"errors"

	"gorm.io/gorm"
)

var (
	ErrPasswordDifferent = errors.New("请设置新密码")
)

type PasswordService interface {
	ModifyPassword(ctx context.Context, username, password string) error
}

type passwordRepository interface {
	GetAccountByUsername(ctx context.Context, username string) (*general.Account, error)
	UpdatePassword(ctx context.Context, account *general.Account, newPassword string) error
}

type defaultPasswordRepository struct {
	db *gorm.DB
}

func (r *defaultPasswordRepository) GetAccountByUsername(ctx context.Context, username string) (*general.Account, error) {
	q := general_query.Use(r.db)
	acc, err := q.Account.WithContext(ctx).Where(
		q.Account.Username.Eq(username),
	).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return acc, nil
}

func (r *defaultPasswordRepository) UpdatePassword(ctx context.Context, account *general.Account, newPassword string) error {
	q := general_query.Use(r.db)
	_, err := q.Account.WithContext(ctx).Where(q.Account.ID.Eq(account.ID)).Update(q.Account.Password, newPassword)
	return err
}

type passwordService struct {
	repo passwordRepository
}

func NewPasswordService(db *gorm.DB) PasswordService {
	return &passwordService{
		repo: &defaultPasswordRepository{db: db},
	}
}

func NewMockPasswordService(repo passwordRepository) PasswordService {
	return &passwordService{
		repo: repo,
	}
}

func (s *passwordService) ModifyPassword(ctx context.Context, username, password string) error {
	// 1. 在 account 表中查找是否有 username 字段值为传入用户名的记录
	account, err := s.repo.GetAccountByUsername(ctx, username)
	if err != nil {
		return err
	}
	if account == nil {
		return ErrUserNotFound // 复用已有的 "该用户不存在"
	}

	// 2. 若密码相同
	if account.Password != nil && *account.Password == password {
		// 则修改该记录的 password 字段值为传入的密码
		err = s.repo.UpdatePassword(ctx, account, password)
		if err != nil {
			return err
		}
		return nil // 修改成功
	}

	// 3. 若密码不同
	return ErrPasswordDifferent
}
