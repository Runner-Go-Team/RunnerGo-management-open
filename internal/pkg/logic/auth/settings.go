package auth

import (
	"context"
	"kp-management/internal/pkg/dal"
	"kp-management/internal/pkg/dal/model"
	"kp-management/internal/pkg/dal/query"
	"kp-management/internal/pkg/dal/rao"
	"kp-management/internal/pkg/packer"
)

func SetUserSettings(ctx context.Context, userID string, settings *rao.UserSettings) error {
	currentTeamID := settings.CurrentTeamID
	tx := query.Use(dal.DB()).Setting
	_, err := tx.WithContext(ctx).Where(tx.UserID.Eq(userID)).UpdateColumnSimple(tx.TeamID.Value(currentTeamID))
	if err != nil {
		return err
	}

	return nil
}

func GetUserSettings(ctx context.Context, userID string) (*rao.GetUserSettingsResp, error) {

	tx := query.Use(dal.DB()).Setting
	settingInfo, err := tx.WithContext(ctx).Where(tx.UserID.Eq(userID)).First()
	if err != nil {
		return nil, err
	}

	userInfo := new(model.User)

	// 查询当前用户在默认团队的角色
	userTeamTable := dal.GetQuery().UserTeam
	utInfo, err := userTeamTable.WithContext(ctx).Where(userTeamTable.TeamID.Eq(settingInfo.TeamID),
		userTeamTable.UserID.Eq(userID)).First()
	if err != nil {
		return nil, err
	}

	// 查询用户信息
	userTable := query.Use(dal.DB()).User
	userInfo, err = userTable.WithContext(ctx).Where(userTable.UserID.Eq(userID)).First()
	if err != nil {
		return nil, err
	}

	return packer.TransUserSettingsToRaoUserSettings(settingInfo, utInfo, userInfo), nil
}

// GetAvailTeamID 获取有效的团队ID
func GetAvailTeamID(ctx context.Context, userID string) (string, error) {
	//获取用户最后一次使用的团队
	tx := query.Use(dal.DB()).Setting
	s, err := tx.WithContext(ctx).Where(tx.UserID.Eq(userID)).First()
	if err != nil {
		return "", err
	}
	lastOperationTeamID := s.TeamID
	return lastOperationTeamID, nil
}
