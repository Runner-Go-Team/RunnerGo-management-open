package auth

import (
	"context"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/packer"
	"github.com/gin-gonic/gin"
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

func GetUserSettings(ctx *gin.Context, userID string) (*rao.GetUserSettingsResp, error) {

	tx := query.Use(dal.DB()).Setting
	settingInfo, err := tx.WithContext(ctx).Where(tx.UserID.Eq(userID)).First()
	if err != nil {
		// 处理兼容问题：超管默认添加一个默认团队
		c := query.Use(dal.DB()).Company
		company, err := c.WithContext(ctx).First()
		if err != nil {
			return nil, err
		}

		r := query.Use(dal.DB()).Role
		superManagerRole, err := r.WithContext(ctx).Where(r.Level.Eq(consts.RoleLevelSuperManager), r.RoleType.Eq(consts.RoleTypeCompany)).First()
		if err != nil {
			return nil, err
		}
		ur := query.Use(dal.DB()).UserRole
		userRole, err := ur.WithContext(ctx).Where(ur.UserID.Eq(userID), ur.CompanyID.Eq(company.CompanyID)).First()
		if err != nil {
			return nil, err
		}
		if userRole.RoleID != superManagerRole.RoleID {
			return nil, err
		}

		teamID := ctx.GetHeader("CurrentTeamID")
		s := query.Use(dal.DB()).Setting
		if err = s.WithContext(ctx).Create(&model.Setting{UserID: userID, TeamID: teamID}); err != nil {
			return nil, err
		}

		settingInfo, err = tx.WithContext(ctx).Where(tx.UserID.Eq(userID)).First()
		if err != nil {
			return nil, err
		}
	}

	userInfo := new(model.User)

	// 查询用户信息
	userTable := query.Use(dal.DB()).User
	userInfo, err = userTable.WithContext(ctx).Where(userTable.UserID.Eq(userID)).First()
	if err != nil {
		return nil, err
	}

	// 查询用户在当前团队内的role_id
	userRoleTB := dal.GetQuery().UserRole
	userTeamRole, err := userRoleTB.WithContext(ctx).Where(userRoleTB.TeamID.Eq(settingInfo.TeamID),
		userRoleTB.UserID.Eq(userID)).First()
	if err != nil {
		return nil, err
	}

	uc := dal.GetQuery().UserCompany
	company, err := uc.WithContext(ctx).Where(uc.UserID.Eq(userID)).First()
	if err != nil {
		return nil, err
	}

	userCompanyRole, err := userRoleTB.WithContext(ctx).Where(userRoleTB.CompanyID.Eq(company.CompanyID),
		userRoleTB.UserID.Eq(userID)).First()
	if err != nil {
		return nil, err
	}

	// 查询角色名称
	roleTB := dal.GetQuery().Role
	roles, err := roleTB.WithContext(ctx).Where(roleTB.CompanyID.Eq(company.CompanyID)).Find()
	if err != nil {
		return nil, err
	}

	return packer.TransUserSettingsToRaoUserSettings(settingInfo, userInfo, userTeamRole, userCompanyRole, roles), nil
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
