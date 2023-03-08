package packer

import (
	"kp-management/internal/pkg/dal/model"
	"kp-management/internal/pkg/dal/rao"
)

func TransEnvDataToRaoEnvList(envData []*model.TeamEnv, envServiceList []rao.ServiceListResp) []rao.EnvListResp {

	resp := make([]rao.EnvListResp, 0, len(envData))
	for _, detail := range envData {

		var serviceListResp []rao.ServiceListResp

		serviceListResp = GetServiceListByEnvID(envServiceList, detail.ID)

		resp = append(resp, rao.EnvListResp{
			ID:          detail.ID,
			Name:        detail.Name,
			Sort:        detail.Sort,
			TeamID:      detail.TeamID,
			ServiceList: serviceListResp,
		})

	}

	return resp
}

func GetServiceListByEnvID(envServiceList []rao.ServiceListResp, EnvID int64) []rao.ServiceListResp {

	var serviceListResp []rao.ServiceListResp

	for _, envServiceListV := range envServiceList {
		if envServiceListV.TeamEnvID == EnvID {
			serviceListResp = append(serviceListResp, rao.ServiceListResp{
				ID:        envServiceListV.ID,
				TeamEnvID: envServiceListV.TeamEnvID,
				TeamID:    envServiceListV.TeamID,
				Name:      envServiceListV.Name,
				Content:   envServiceListV.Content,
			})
		}
	}

	return serviceListResp
}
