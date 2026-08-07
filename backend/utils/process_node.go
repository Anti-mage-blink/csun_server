package utils

import (
	"context"
	"errors"
	"fmt"

	"csun_server-backend/dao/model/general"
	general_query "csun_server-backend/dao/query/general"

	"github.com/expr-lang/expr"
	"gorm.io/gorm"
)

// QueryNextNodeNum 查询下一节点编号（作为工具函数）
// 函数参数：当前节点序号 present_node_num，条件参数名 cond_param_name、条件参数值 cond_param_value（后两者可为空）
// 返回：next_node（一条节点记录，即 general.process_node）
func QueryNextNodeNum(ctx context.Context, db *gorm.DB, presentNodeNum *int32, condParamName string, condParamValue interface{}) (*general.ProcessNode, error) {
	qGen := general_query.Use(db)

	// 1. 用当前节点序号（quote_manage.quote_process.present_num）作为 start_node_num
	//（若为空或0，则下一节点为 node_num==1 的节点，返回该节点记录）
	if presentNodeNum == nil || *presentNodeNum == 0 {
		firstNode, err := qGen.ProcessNode.WithContext(ctx).
			Where(qGen.ProcessNode.NodeNum.Eq(1)).
			First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil
			}
			return nil, fmt.Errorf("查询 node_num=1 的节点记录失败: %w", err)
		}
		return firstNode, nil
	}

	// 查询边 general.process_edge
	edges, err := qGen.ProcessEdge.WithContext(ctx).
		Where(qGen.ProcessEdge.StartNodeNum.Eq(*presentNodeNum)).
		Find()
	if err != nil {
		return nil, fmt.Errorf("查询流程边失败: %w", err)
	}

	if len(edges) == 0 {
		return nil, nil
	}

	var targetNodeNum *int32

	// 2. 判断边类型
	isConditionType := false
	if edges[0].Type != nil && *edges[0].Type == "condition" {
		isConditionType = true
	}

	if !isConditionType {
		// 若类型 type 不为 condition，拿到下一个一般节点 general.process_edge.end_node_num，用它作为 node_num 去查
		targetNodeNum = edges[0].EndNodeNum
	} else {
		// 若类型 type 为 condition（查到多条边），则对多条边逐条判定记录.condition_expr（用 Go 的 expr 等 Go 中针对该问题专用成熟标准库），判定成功命中为止
		env := make(map[string]interface{})
		if condParamName != "" {
			env[condParamName] = condParamValue
		}

		for _, edge := range edges {
			if edge.ConditionExpr == nil || *edge.ConditionExpr == "" {
				continue
			}

			program, compErr := expr.Compile(*edge.ConditionExpr, expr.Env(env))
			if compErr != nil {
				continue
			}

			output, runErr := expr.Run(program, env)
			if runErr != nil {
				continue
			}

			if boolVal, ok := output.(bool); ok && boolVal {
				// 对命中的边，拿到下一个一般节点记录
				targetNodeNum = edge.EndNodeNum
				break
			}
		}
	}

	if targetNodeNum == nil {
		// 若不成功，则没有下一个一般节点记录，返回空
		return nil, nil
	}

	// 3. 拿到 end_node_num 后，用它作为 node_num 去查 general.process_node
	targetNode, err := qGen.ProcessNode.WithContext(ctx).
		Where(qGen.ProcessNode.NodeNum.Eq(*targetNodeNum)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("查询一般流程节点失败: %w", err)
	}

	return targetNode, nil
}
