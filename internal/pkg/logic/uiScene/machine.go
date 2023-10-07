package uiScene

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/clients"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/gin-gonic/gin"
	"strings"
)

// GetUiEngineMachineList 获取机器列表
func GetUiEngineMachineList(ctx *gin.Context, req *rao.UISceneEngineMachineReq) ([]*rao.UiEngineMachineInfo, error) {
	list, err := clients.GetUiEngineMachineList()
	if err != nil {
		return nil, err
	}

	result := make([]*rao.UiEngineMachineInfo, 0, len(list))
	keyword := strings.TrimSpace(req.Keyword)
	for _, info := range list {
		if len(keyword) > 0 {
			if info.SystemInfo != nil && strings.Contains(info.SystemInfo.Hostname, keyword) {
				result = append(result, info)
			}
		} else {
			result = append(result, info)
		}
	}

	return result, nil
}
