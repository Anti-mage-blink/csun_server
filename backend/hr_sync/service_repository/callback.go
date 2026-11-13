package service_repository

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	general_service "csun_server-backend/general/service_repository"
)

const (
	EventKeyTest = "event_test"

	ResultCodeSuccess = "SUCCESS"
	ResultCodeFail    = "FAIL"
	ResultMsgOK       = "OK"
)

type CallbackParams struct {
	Key  string
	Data json.RawMessage
}

type CallbackResult struct {
	ResultCode string `json:"result_code"`
	ResultMsg  string `json:"result_msg"`
}

type CallbackService interface {
	HandleCallback(ctx context.Context, params *CallbackParams) (*CallbackResult, error)
}

type callbackRepository interface{}

type defaultCallbackRepository struct {
	db *general_service.DBEngine
}

type callbackService struct {
	repo callbackRepository
}

func NewCallbackService(db *general_service.DBEngine) CallbackService {
	return &callbackService{
		repo: &defaultCallbackRepository{db: db},
	}
}

func NewMockCallbackService(repo callbackRepository) CallbackService {
	return &callbackService{
		repo: repo,
	}
}

func (s *callbackService) HandleCallback(ctx context.Context, params *CallbackParams) (*CallbackResult, error) {
	if params == nil || params.Key == "" {
		return nil, errors.New("回调事件 key 不能为空")
	}

	log.Printf("[HRSync Callback] 收到回调事件: key=%s, data=%s", params.Key, string(params.Data))

	switch params.Key {
	case EventKeyTest:
		return &CallbackResult{
			ResultCode: ResultCodeSuccess,
			ResultMsg:  ResultMsgOK,
		}, nil
	default:
		// 未接入具体同步逻辑前先按规范应答 SUCCESS，避免 2号人事判定回调失败触发重试或停用回调
		log.Printf("[HRSync Callback] 暂未绑定具体业务处理的事件类型: key=%s", params.Key)
		return &CallbackResult{
			ResultCode: ResultCodeSuccess,
			ResultMsg:  ResultMsgOK,
		}, nil
	}
}
