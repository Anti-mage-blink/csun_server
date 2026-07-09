package repository

import (
	"context"
	"errors"
	"csun_server-backend/dao/model/general"
	general_query "csun_server-backend/dao/query/general"
	"gorm.io/gorm"
)

type PasswordRepository interface {
	GetAccountByUsername(ctx context.Context, username string) (*general.Account, error)
	UpdatePassword(ctx context.Context, account *general.Account, newPassword string) error
}

type passwordRepository struct {
	db *gorm.DB
}

func NewPasswordRepository(db *gorm.DB) PasswordRepository {
	return &passwordRepository{db: db}
}

func (r *passwordRepository) GetAccountByUsername(ctx context.Context, username string) (*general.Account, error) {
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

func (r *passwordRepository) UpdatePassword(ctx context.Context, account *general.Account, newPassword string) error {
	q := general_query.Use(r.db)
	_, err := q.Account.WithContext(ctx).Where(q.Account.ID.Eq(account.ID)).Update(q.Account.Password, newPassword)
	return err
}
