package service_repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var nameReg = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// ParseDBAndTable 解析 "数据库名.数据表名" 字符串，一处实现、多处复用
func ParseDBAndTable(fullStr string) (dbName, tableName string, err error) {
	fullStr = strings.TrimSpace(fullStr)
	if fullStr == "" {
		return "", "", errors.New("数据表标识不能为空")
	}

	parts := strings.Split(fullStr, ".")
	if len(parts) == 1 {
		tableName = strings.TrimSpace(parts[0])
	} else if len(parts) == 2 {
		dbName = strings.TrimSpace(parts[0])
		tableName = strings.TrimSpace(parts[1])
	} else {
		return "", "", errors.New("数据表标识格式不正确，应为 '数据表名' 或 '数据库名.数据表名'")
	}

	if tableName == "" {
		return "", "", errors.New("数据表名不能为空")
	}

	if dbName != "" && !nameReg.MatchString(dbName) {
		return "", "", errors.New("不合法的数据库名")
	}
	if !nameReg.MatchString(tableName) {
		return "", "", errors.New("不合法的数据表名")
	}

	return dbName, tableName, nil
}

// getFullTableName 防 SQL 注入校验并构造完整表名
func getFullTableName(dbName, tableName string) (string, error) {
	if tableName == "" {
		return "", errors.New("数据表名不能为空")
	}
	if dbName != "" {
		return fmt.Sprintf("`%s`.`%s`", dbName, tableName), nil
	}
	return fmt.Sprintf("`%s`", tableName), nil
}

// QueryResult 包含主表及关联数据表记录的全量查询结果
type QueryResult struct {
	Main      []map[string]interface{}            `json:"main"`
	Relations map[string][]map[string]interface{} `json:"relations"`
}

// DataTableService 通用数据表 CRUD 服务接口
type DataTableService interface {
	ParseDBAndTable(fullStr string) (dbName, tableName string, err error)
	QueryList(ctx context.Context, tableStr string, relationTables []string) (*QueryResult, error)
	CreateRecord(ctx context.Context, tableStr string, record map[string]interface{}) (map[string]interface{}, error)
	UpdateRecord(ctx context.Context, tableStr string, id interface{}, record map[string]interface{}) error
	DeleteRecord(ctx context.Context, tableStr string, id interface{}) error
}

type defaultDataTableRepository struct {
	db *DBEngine
}

// NewDataTableService 创建 DataTableService 实例
func NewDataTableService(db *DBEngine) DataTableService {
	return &defaultDataTableRepository{db: db}
}

// ParseDBAndTable 挂载到服务接口，供外部复用
func (r *defaultDataTableRepository) ParseDBAndTable(fullStr string) (string, string, error) {
	return ParseDBAndTable(fullStr)
}

// querySingleTable 全量查询单表数据
func (r *defaultDataTableRepository) getTableColumns(ctx context.Context, dbName, tableName string) (map[string]bool, error) {
	var columns []string
	var err error
	if dbName != "" {
		err = r.db.General.WithContext(ctx).Raw(
			"SELECT COLUMN_NAME FROM information_schema.columns WHERE table_schema = ? AND table_name = ?", dbName, tableName,
		).Scan(&columns).Error
	} else {
		err = r.db.General.WithContext(ctx).Raw(
			"SELECT COLUMN_NAME FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ?", tableName,
		).Scan(&columns).Error
	}
	if err != nil {
		return nil, err
	}
	colsMap := make(map[string]bool, len(columns))
	for _, col := range columns {
		colsMap[strings.ToLower(col)] = true
	}
	return colsMap, nil
}

// querySingleTable 全量查询单表数据
func (r *defaultDataTableRepository) querySingleTable(ctx context.Context, dbName, tableName string) ([]map[string]interface{}, error) {
	fullTable, err := getFullTableName(dbName, tableName)
	if err != nil {
		return nil, err
	}

	colsMap, err := r.getTableColumns(ctx, dbName, tableName)
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	db := r.db.General.WithContext(ctx).Table(fullTable)

	// 1. 若表存在 is_deleted 字段，独立过滤软删除标识记录
	if colsMap["is_deleted"] {
		db = db.Where("is_deleted = ? OR is_deleted IS NULL OR is_deleted = 0", false)
	}

	// 2. 若表存在 id 字段，按 id DESC 排序
	if colsMap["id"] {
		db = db.Order("id DESC")
	}

	err = db.Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

// QueryList 查：返回该数据表及所有关联数据表的全量记录列表
func (r *defaultDataTableRepository) QueryList(ctx context.Context, tableStr string, relationTables []string) (*QueryResult, error) {
	dbName, tableName, err := ParseDBAndTable(tableStr)
	if err != nil {
		return nil, err
	}

	mainList, err := r.querySingleTable(ctx, dbName, tableName)
	if err != nil {
		return nil, fmt.Errorf("查询主数据表 [%s] 失败: %w", tableStr, err)
	}

	relationsMap := make(map[string][]map[string]interface{})
	for _, relStr := range relationTables {
		relStr = strings.TrimSpace(relStr)
		if relStr == "" {
			continue
		}

		relDb, relTable, parseErr := ParseDBAndTable(relStr)
		if parseErr != nil {
			return nil, fmt.Errorf("解析关联数据表 [%s] 失败: %w", relStr, parseErr)
		}

		relList, queryErr := r.querySingleTable(ctx, relDb, relTable)
		if queryErr != nil {
			return nil, fmt.Errorf("查询关联数据表 [%s] 失败: %w", relStr, queryErr)
		}

		relationsMap[relStr] = relList
	}

	return &QueryResult{
		Main:      mainList,
		Relations: relationsMap,
	}, nil
}

// CreateRecord 增：写入数据记录 (id 自增)
func (r *defaultDataTableRepository) CreateRecord(ctx context.Context, tableStr string, record map[string]interface{}) (map[string]interface{}, error) {
	dbName, tableName, err := ParseDBAndTable(tableStr)
	if err != nil {
		return nil, err
	}

	fullTable, err := getFullTableName(dbName, tableName)
	if err != nil {
		return nil, err
	}

	colsMap, err := r.getTableColumns(ctx, dbName, tableName)
	if err != nil {
		return nil, err
	}

	if record == nil {
		record = make(map[string]interface{})
	}

	// 移除空的 id 字段，交由数据库主键自增生成
	if idVal, exists := record["id"]; exists {
		if idVal == nil || idVal == "" || idVal == float64(0) || idVal == int64(0) || idVal == 0 {
			delete(record, "id")
		}
	}

	// 处理空字符串，转为 nil 以规避数字/浮点字段接受空字符串 "" 时报 SQL 错误；同时移除不属于该表的多余字段
	for k, v := range record {
		if strVal, ok := v.(string); ok && strVal == "" {
			record[k] = nil
		}
		if len(colsMap) > 0 && !colsMap[strings.ToLower(k)] {
			delete(record, k)
		}
	}

	// 设置默认时间戳
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	if colsMap["created_at"] {
		if val, exists := record["created_at"]; !exists || val == nil || val == "" {
			record["created_at"] = nowStr
		}
	} else {
		delete(record, "created_at")
	}

	if colsMap["updated_at"] {
		if val, exists := record["updated_at"]; !exists || val == nil || val == "" {
			record["updated_at"] = nowStr
		}
	} else {
		delete(record, "updated_at")
	}

	err = r.db.General.WithContext(ctx).Table(fullTable).Create(&record).Error
	if err != nil {
		return nil, err
	}
	return record, nil
}

// UpdateRecord 改：根据 id 修改数据记录
func (r *defaultDataTableRepository) UpdateRecord(ctx context.Context, tableStr string, id interface{}, record map[string]interface{}) error {
	dbName, tableName, err := ParseDBAndTable(tableStr)
	if err != nil {
		return err
	}

	fullTable, err := getFullTableName(dbName, tableName)
	if err != nil {
		return err
	}

	if id == nil || id == "" || id == 0 {
		return errors.New("修改操作必须指定有效的记录 ID")
	}

	if record == nil {
		return errors.New("修改记录的提交内容不能为空")
	}

	colsMap, err := r.getTableColumns(ctx, dbName, tableName)
	if err != nil {
		return err
	}

	// 保护主键 ID 不被覆盖
	delete(record, "id")

	// 处理空字符串，转为 nil 以规避数字/浮点字段接受空字符串 "" 时报 SQL 错误；同时移除不属于该表的多余字段
	for k, v := range record {
		if strVal, ok := v.(string); ok && strVal == "" {
			record[k] = nil
		}
		if len(colsMap) > 0 && !colsMap[strings.ToLower(k)] {
			delete(record, k)
		}
	}

	if colsMap["updated_at"] {
		nowStr := time.Now().Format("2006-01-02 15:04:05")
		record["updated_at"] = nowStr
	} else {
		delete(record, "updated_at")
	}

	if !colsMap["created_at"] {
		delete(record, "created_at")
	}

	return r.db.General.WithContext(ctx).Table(fullTable).Where("id = ?", id).Updates(record).Error
}

// DeleteRecord 删：根据 id 软/硬删除数据记录
func (r *defaultDataTableRepository) DeleteRecord(ctx context.Context, tableStr string, id interface{}) error {
	dbName, tableName, err := ParseDBAndTable(tableStr)
	if err != nil {
		return err
	}

	fullTable, err := getFullTableName(dbName, tableName)
	if err != nil {
		return err
	}

	if id == nil || id == "" || id == 0 {
		return errors.New("删除操作必须指定有效的记录 ID")
	}

	colsMap, err := r.getTableColumns(ctx, dbName, tableName)
	if err != nil {
		return err
	}

	// 独立判断是否存在 is_deleted 字段
	if colsMap["is_deleted"] {
		updates := map[string]interface{}{
			"is_deleted": true,
		}
		// 独立判断是否存在 updated_at 字段
		if colsMap["updated_at"] {
			nowStr := time.Now().Format("2006-01-02 15:04:05")
			updates["updated_at"] = nowStr
		}
		return r.db.General.WithContext(ctx).Table(fullTable).Where("id = ?", id).Updates(updates).Error
	}

	// 若数据表无 is_deleted 字段，则执行物理硬删除
	return r.db.General.WithContext(ctx).Table(fullTable).Where("id = ?", id).Delete(nil).Error
}
