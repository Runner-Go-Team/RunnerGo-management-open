package middleware

import (
	"errors"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/jwt"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusOK, gin.H{"code": errno.ErrMustLogin, "em": "must login", "et": "需要登录", "data": ""})
			c.Abort()
			return
		}

		userID, err := jwt.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": errno.ErrMustLogin, "em": "must login", "et": "需要登录", "data": ""})
			c.Abort()
			return
		} else if userID == "" {
			c.JSON(http.StatusOK, gin.H{"code": errno.ErrMustLogin, "em": "must login", "et": "需要登录", "data": ""})
			c.Abort()
			return
		}

		// 用户是否需要重新登录
		if exists, _ := dal.GetRDB().SIsMember(c, consts.RedisResetLoginUsers, userID).Result(); exists {
			c.JSON(http.StatusOK, gin.H{"code": errno.ErrAccountDel, "em": "reset login", "et": "请重新登录"})
			c.Abort()
			return
		}

		// 查询token里面的用户信息是否存在于数据库
		userTable := dal.GetQuery().User
		_, err = userTable.WithContext(c).Where(userTable.UserID.Eq(userID)).First()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, gin.H{"code": errno.ErrAccountDel, "em": "user not found", "et": "用户不存在或已删除"})
			c.Abort()
			return
		}
		if err != nil {
			// 把token设置为过期
			_, _, err := jwt.GenerateTokenByTime(userID, 0)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"code": errno.ErrInvalidToken, "em": "must login", "et": "需要登录", "data": ""})
				c.Abort()
				return
			}

			c.JSON(http.StatusOK, gin.H{"code": errno.ErrMustLogin, "em": "must login", "et": "需要登录", "data": ""})
			c.Abort()
			return
		}

		// 过滤部分接口的校验条件
		apiPath := c.Request.URL.Path
		if apiPath != "/management/api/v1/setting/get" && apiPath != "/management/api/v1/setting/set" {
			// 查询当前用户是否存在于当前团队
			teamID := c.GetHeader("CurrentTeamID")
			userTeamTable := dal.GetQuery().UserTeam
			setDefaultTeamErr := false
			if teamID != "" {
				_, err = userTeamTable.WithContext(c).Where(userTeamTable.TeamID.Eq(teamID), userTeamTable.UserID.Eq(userID)).First()
				if err != nil {
					setDefaultTeamErr = true
				}
			} else {
				// 查询当前用户最新加入的一个团队
				utInfo, err := userTeamTable.WithContext(c).Where(userTeamTable.UserID.Eq(userID),
					userTeamTable.IsShow.Eq(consts.TeamIsShow)).Order(userTeamTable.CreatedAt.Desc()).First()
				if err != nil {
					setDefaultTeamErr = true
				} else {
					// 设置一个默认团队
					setTB := dal.GetQuery().Setting
					_, err = setTB.WithContext(c).Where(setTB.UserID.Eq(userID)).Assign(
						setTB.TeamID.Value(utInfo.TeamID)).FirstOrCreate()
					if err != nil {
						setDefaultTeamErr = true
					}
				}
			}

			if setDefaultTeamErr == true {
				c.JSON(http.StatusOK, gin.H{"code": errno.ErrDefaultTeamFailed, "em": "default team failed",
					"et": "查询的默认团队错误", "data": ""})
				c.Abort()
				return
			}
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
