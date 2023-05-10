package packer

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
)

func TransOperationsToRaoOperationList(operations []*mao.OperationLog, users []*model.User) []*rao.Operation {
	ret := make([]*rao.Operation, 0)

	memo := make(map[string]*model.User)
	for _, user := range users {
		memo[user.UserID] = user
	}

	for _, operationInfo := range operations {
		ret = append(ret, &rao.Operation{
			UserID:         operationInfo.UserID,
			UserName:       memo[operationInfo.UserID].Nickname,
			UserAvatar:     memo[operationInfo.UserID].Avatar,
			UserStatus:     0,
			Category:       operationInfo.Category,
			Operate:        operationInfo.Operate,
			Name:           operationInfo.Name,
			CreatedTimeSec: operationInfo.CreatedAt,
		})
	}

	return ret
}
