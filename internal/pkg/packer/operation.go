package packer

import (
	"kp-management/internal/pkg/dal/mao"
	"kp-management/internal/pkg/dal/model"
	"kp-management/internal/pkg/dal/rao"
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
