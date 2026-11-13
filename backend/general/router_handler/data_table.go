package router_handler

import (
	"net/http"
	"strings"

	"csun_server-backend/service_repository"

	"github.com/gin-gonic/gin"
)

// DataTableQueryRequest 查接口请求绑定结构体
type DataTableQueryRequest struct {
	Table          string   `json:"table" form:"table"`
	DBTable        string   `json:"db_table" form:"db_table"`
	TableName      string   `json:"table_name" form:"table_name"`
	DBName         string   `json:"db_name" form:"db_name"`
	RelationTables []string `json:"relation_tables" form:"relation_tables"`
}

// DataTableMutateRequest 增/改/删接口请求绑定结构体
type DataTableMutateRequest struct {
	Table     string                 `json:"table" form:"table"`
	DBTable   string                 `json:"db_table" form:"db_table"`
	TableName string                 `json:"table_name" form:"table_name"`
	DBName    string                 `json:"db_name" form:"db_name"`
	ID        interface{}            `json:"id" form:"id"`
	Record    map[string]interface{} `json:"record"`
	Data      map[string]interface{} `json:"data"`
}

// DataTableHandler 通用数据表 Handler
type DataTableHandler struct {
	service service_repository.DataTableService
}

// NewDataTableHandler 创建 Handler 实例
func NewDataTableHandler(service service_repository.DataTableService) *DataTableHandler {
	return &DataTableHandler{service: service}
}

// RegisterDataTableRoutes 注册通用数据表相关的 4 种 CRUD 路由
func RegisterDataTableRoutes(r *gin.Engine, h *DataTableHandler) {
	api := r.Group("/api")
	{
		api.GET("/data_table", h.Query)
		api.POST("/data_table", h.Create)
		api.PUT("/data_table", h.Update)
		api.DELETE("/data_table", h.Delete)
	}
}

// extractTableStr 统一合成/抽取 "数据库名.数据表名" 字符串
func extractTableStr(table, dbTable, tableName, dbName string) string {
	table = strings.TrimSpace(table)
	if table != "" {
		return table
	}
	dbTable = strings.TrimSpace(dbTable)
	if dbTable != "" {
		return dbTable
	}
	tableName = strings.TrimSpace(tableName)
	dbName = strings.TrimSpace(dbName)
	if dbName != "" && tableName != "" {
		return dbName + "." + tableName
	}
	return tableName
}

// Query 查：返回该数据表及所有关联数据表的全量记录列表
func (h *DataTableHandler) Query(c *gin.Context) {
	var req DataTableQueryRequest
	_ = c.ShouldBindQuery(&req)

	tableStr := extractTableStr(req.Table, req.DBTable, req.TableName, req.DBName)
	if tableStr == "" {
		tableStr = extractTableStr(c.Query("table"), c.Query("db_table"), c.Query("table_name"), c.Query("db_name"))
	}

	if tableStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "缺少必需的数据表标识参数 (table 或 db_name.table_name)"})
		return
	}

	relationTables := req.RelationTables
	if len(relationTables) == 0 {
		if relStr := c.Query("relation_tables"); relStr != "" {
			relationTables = strings.Split(relStr, ",")
		}
	}

	result, err := h.service.QueryList(c.Request.Context(), tableStr, relationTables)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询数据表失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "查询数据表成功",
		"data":      result.Main,
		"relations": result.Relations,
	})
}

// Create 增：新增一条记录
func (h *DataTableHandler) Create(c *gin.Context) {
	var req DataTableMutateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数解析错误: " + err.Error()})
		return
	}

	tableStr := extractTableStr(req.Table, req.DBTable, req.TableName, req.DBName)
	if tableStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "缺少必需的数据表标识参数 (table 或 db_name.table_name)"})
		return
	}

	record := req.Record
	if record == nil {
		record = req.Data
	}

	createdRecord, err := h.service.CreateRecord(c.Request.Context(), tableStr, record)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "新增记录失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "新增记录成功",
		"data":    createdRecord,
	})
}

// Update 改：根据 ID 修改记录
func (h *DataTableHandler) Update(c *gin.Context) {
	var req DataTableMutateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数解析错误: " + err.Error()})
		return
	}

	tableStr := extractTableStr(req.Table, req.DBTable, req.TableName, req.DBName)
	if tableStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "缺少必需的数据表标识参数 (table 或 db_name.table_name)"})
		return
	}

	id := req.ID
	if id == nil || id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "缺少必需参数 id"})
		return
	}

	record := req.Record
	if record == nil {
		record = req.Data
	}

	err := h.service.UpdateRecord(c.Request.Context(), tableStr, id, record)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "修改记录失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "修改记录成功",
	})
}

// Delete 删：根据 ID 软删除记录
func (h *DataTableHandler) Delete(c *gin.Context) {
	var req DataTableMutateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求参数解析错误: " + err.Error()})
		return
	}

	tableStr := extractTableStr(req.Table, req.DBTable, req.TableName, req.DBName)
	if tableStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "缺少必需的数据表标识参数 (table 或 db_name.table_name)"})
		return
	}

	id := req.ID
	if id == nil || id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "缺少必需参数 id"})
		return
	}

	err := h.service.DeleteRecord(c.Request.Context(), tableStr, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "删除记录失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除记录成功",
	})
}
