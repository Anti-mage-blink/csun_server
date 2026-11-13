package service_repository

import (
	"context"
	"errors"
	"log"

	"csun_server-backend/dao/model/general"
	"csun_server-backend/dao/model/quote_manage"
	general_query "csun_server-backend/dao/query/general"
	quote_query "csun_server-backend/dao/query/quote_manage"

	"gorm.io/gorm"
)

var (
	ErrPasswordIncorrect = errors.New("密码错误")
	ErrUserNotFound      = errors.New("该用户不存在")
	ErrEmployeeNotFound  = errors.New("关联的员工不存在")
)

type LoginUser struct {
	ID             int32   `json:"id"`
	EmployeeNumber *string `json:"employee_number"`
	Name           *string `json:"name"`
	Department     *string `json:"department"`
	Role           string  `json:"role"`
}

type AuthService interface {
	Login(ctx context.Context, username, password string) (*LoginUser, error)
}

// authRepository 接口隐藏在 service_repository 层内部，便于单元测试 mock
type authRepository interface {
	GetAccountByUsername(ctx context.Context, username string) (*general.Account, error)
	GetEmployeeByID(ctx context.Context, employeeID int32) (*general.Employee, error)
	GetEmployeeRoleByEmployeeID(ctx context.Context, employeeID int32) (*quote_manage.EmployeeRole, error)
}

type defaultAuthRepository struct {
	db *gorm.DB
}

func (r *defaultAuthRepository) GetAccountByUsername(ctx context.Context, username string) (*general.Account, error) {
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

func (r *defaultAuthRepository) GetEmployeeByID(ctx context.Context, employeeID int32) (*general.Employee, error) {
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

func (r *defaultAuthRepository) GetEmployeeRoleByEmployeeID(ctx context.Context, employeeID int32) (*quote_manage.EmployeeRole, error) {
	q := quote_query.Use(r.db)
	role, err := q.EmployeeRole.WithContext(ctx).Where(
		q.EmployeeRole.EmployeeID.Eq(employeeID),
	).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return role, nil
}

type authService struct {
	repo authRepository
}

func NewAuthService(db *gorm.DB) AuthService {
	return &authService{
		repo: &defaultAuthRepository{db: db},
	}
}

// NewMockAuthService 用于在需要时注入 Mock Repository (主要在测试中使用)
func NewMockAuthService(repo authRepository) AuthService {
	return &authService{
		repo: repo,
	}
}

func (s *authService) Login(ctx context.Context, username, password string) (*LoginUser, error) {
	// 1. 查找是否有该 username
	account, err := s.repo.GetAccountByUsername(ctx, username)
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

	employee, err := s.repo.GetEmployeeByID(ctx, *account.EmployeeID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	// 4. 获取员工角色信息
	var roleStr = "普通员工"
	empRole, err := s.repo.GetEmployeeRoleByEmployeeID(ctx, *account.EmployeeID)
	if err != nil {
		log.Printf("[Login] GetEmployeeRoleByEmployeeID 发生错误: employee_id=%d, err=%v", *account.EmployeeID, err)
	} else if empRole == nil {
		log.Printf("[Login] GetEmployeeRoleByEmployeeID 结果为 nil: employee_id=%d", *account.EmployeeID)
	} else {
		if empRole.Role != nil {
			roleStr = *empRole.Role
			log.Printf("[Login] GetEmployeeRoleByEmployeeID 成功获取角色: employee_id=%d, role=%s", *account.EmployeeID, roleStr)
		} else {
			log.Printf("[Login] GetEmployeeRoleByEmployeeID 记录中 role 为空: employee_id=%d", *account.EmployeeID)
		}
	}

	return &LoginUser{
		ID:             employee.ID,
		EmployeeNumber: employee.EmployeeNumber,
		Name:           employee.Name,
		Department:     employee.Department,
		Role:           roleStr,
	}, nil
}
