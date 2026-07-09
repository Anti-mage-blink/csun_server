package repository

import (
	"context"
	"errors"
	"csun_server-backend/dao/model/general"
	general_query "csun_server-backend/dao/query/general"
	"gorm.io/gorm"
)

type AuthRepository interface {
	GetAccountByUsername(ctx context.Context, username string) (*general.Account, error)
	GetEmployeeByID(ctx context.Context, employeeID int32) (*general.Employee, error)
}

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) GetAccountByUsername(ctx context.Context, username string) (*general.Account, error) {
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

func (r *authRepository) GetEmployeeByID(ctx context.Context, employeeID int32) (*general.Employee, error) {
	q := general_query.Use(r.db)
	emp, err := q.Employee.WithContext(ctx).Where(
		q.Employee.ID.Eq(employeeID),
	).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return emp, nil
}
