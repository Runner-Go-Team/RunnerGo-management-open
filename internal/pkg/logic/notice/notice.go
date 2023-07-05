package notice

import (
	"context"
	"encoding/json"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/mail"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/query"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/logic/permission"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/public"
	"github.com/gin-gonic/gin"
)

// SaveNoticeEvent 保存通知事件
func SaveNoticeEvent(ctx context.Context, userID string, req *rao.SaveNoticeEventReq) error {
	planIDs := req.PlanIDs
	groupIDs := req.GroupIDs

	err := query.Use(dal.DB()).Transaction(func(tx *query.Query) error {
		// 删除
		if _, err := tx.ThirdNoticeGroupEvent.WithContext(ctx).Where(
			tx.ThirdNoticeGroupEvent.EventID.Eq(req.EventID),
			tx.ThirdNoticeGroupEvent.TeamID.Eq(req.TeamID),
			tx.ThirdNoticeGroupEvent.PlanID.In(planIDs...),
		).Delete(); err != nil {
			return err
		}

		// 添加
		events := make([]*model.ThirdNoticeGroupEvent, 0)
		for _, groupID := range groupIDs {
			for _, planID := range planIDs {
				events = append(events, &model.ThirdNoticeGroupEvent{
					GroupID: groupID,
					EventID: req.EventID,
					PlanID:  planID,
					TeamID:  req.TeamID,
				})
			}
		}
		if err := tx.ThirdNoticeGroupEvent.WithContext(ctx).Create(events...); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// GetGroupNoticeEvent 获取组通知事件
func GetGroupNoticeEvent(ctx *gin.Context, req *rao.GetGroupNoticeEventReq) ([]string, error) {
	nge := dal.GetQuery().ThirdNoticeGroupEvent

	var groupIDs []string
	if err := nge.WithContext(ctx).Where(
		nge.EventID.Eq(req.EventID),
		nge.TeamID.Eq(req.TeamID),
		nge.PlanID.Eq(req.PlanID),
	).Pluck(nge.GroupID, &groupIDs); err != nil {
		return nil, err
	}

	return groupIDs, nil
}

// SendNoticeByGroup 通过组ID 发送通知
func SendNoticeByGroup(ctx *gin.Context, groupID string, params *rao.SendCardParams) error {
	ngr := dal.GetQuery().ThirdNoticeGroupRelate
	noticeGroupRelate, err := ngr.WithContext(ctx).Where(ngr.GroupID.Eq(groupID)).Find()
	if err != nil {
		return err
	}

	noticeIDs := make([]string, 0, len(noticeGroupRelate))
	for _, relate := range noticeGroupRelate {
		noticeIDs = append(noticeIDs, relate.NoticeID)
	}

	tn := dal.GetQuery().ThirdNotice
	notices, err := tn.WithContext(ctx).Where(tn.NoticeID.In(noticeIDs...), tn.Status.Eq(consts.NoticeStatusNormal)).Find()
	if err != nil {
		return err
	}

	noticeMemo := make(map[string]*model.ThirdNotice)
	for _, notice := range notices {
		noticeMemo[notice.NoticeID] = notice
	}

	for _, nr := range noticeGroupRelate {
		n, ok := noticeMemo[nr.NoticeID]
		if !ok {
			continue
		}

		switch n.ChannelID {
		case consts.NoticeChannelIDFRobot:
			s := &rao.FeiShuRobot{}
			_ = json.Unmarshal([]byte(n.Params), s)
			if err := SendFeiShuBot(ctx, s, getCard(params)); err != nil {
				return err
			}
		case consts.NoticeChannelIDFApp:
			s := &rao.FeiShuApp{}
			_ = json.Unmarshal([]byte(n.Params), &s)

			relateParams := &rao.NoticeGroupRelateParams{}
			_ = json.Unmarshal([]byte(nr.Params), &relateParams)

			// 过滤用户
			userIDs, _ := FilterUserByThird(ctx, n.NoticeID, relateParams.UserIDs)
			if len(userIDs) <= 0 {
				return nil
			}
			if err := SendFeiShuApp(ctx, s, userIDs, getCardContent(params)); err != nil {
				return err
			}
		case consts.NoticeChannelIDWxRobot:
			s := &rao.WechatRobot{}
			_ = json.Unmarshal([]byte(n.Params), s)
			if err := SendWechatBot(ctx, s, getWechatCardContent(params)); err != nil {
				return err
			}
		case consts.NoticeChannelIDEmail:
			s := &rao.SMTPEmail{}
			_ = json.Unmarshal([]byte(n.Params), &s)

			relateParams := &rao.NoticeGroupRelateParams{}
			_ = json.Unmarshal([]byte(nr.Params), &relateParams)

			if len(relateParams.Emails) <= 0 {
				return nil
			}
			for _, email := range relateParams.Emails {
				if params.ReportType == consts.PlanStress {
					if err := mail.SendPlanNoticeEmail(ctx, s, email, params); err != nil {
						return err
					}
				}
				if params.ReportType == consts.PlanAuto {
					if err := mail.SendAutoPlanNoticeEmail(ctx, s, email, params); err != nil {
						return err
					}
				}
			}
		case consts.NoticeChannelIDDingRobot:
			s := &rao.DingTalkRobot{}
			_ = json.Unmarshal([]byte(n.Params), s)
			if err := SendDingTalkBot(ctx, s, getDingCardContent(params)); err != nil {
				return err
			}
		case consts.NoticeChannelIDDingApp:
			s := &rao.DingTalkApp{}
			_ = json.Unmarshal([]byte(n.Params), &s)

			relateParams := &rao.NoticeGroupRelateParams{}
			_ = json.Unmarshal([]byte(nr.Params), &relateParams)

			userIDs, _ := FilterUserByThird(ctx, n.NoticeID, relateParams.UserIDs)
			if len(userIDs) <= 0 {
				return nil
			}
			if err := SendDingTalkApp(ctx, s, userIDs, getDingAppCardContent(params)); err != nil {
				return err
			}
		}
	}

	return nil
}

func FindOpenIDs(users []rao.ThirdUserInfo) []string {
	var openIDs []string
	for _, user := range users {
		if user.OpenID != "" {
			openIDs = append(openIDs, user.OpenID)
		}
	}
	return openIDs
}

func RecursiveFindOpenIDs(departments []rao.ThirdDepartmentInfo) []string {
	var openIDs []string
	for _, dept := range departments {
		openIDs = append(openIDs, FindOpenIDs(dept.UserList)...)
		openIDs = append(openIDs, RecursiveFindOpenIDs(dept.DepartmentList)...)
	}
	return openIDs
}

func FilterUserByThird(ctx *gin.Context, noticeID string, userIDs []string) ([]string, error) {
	getNoticeThirdUsersReq := &rao.GetNoticeThirdUsersReq{NoticeID: noticeID}
	users, err := permission.GetNoticeThirdUsers(ctx, getNoticeThirdUsersReq)
	if err != nil {
		return userIDs, err
	}

	departmentLists := &rao.ThirdCompanyUsers{}
	// 将接口类型转换为结构体类型
	departmentList, err := json.Marshal(users)
	if err != nil {
		return userIDs, err
	}
	if err := json.Unmarshal(departmentList, departmentLists); err != nil {
		return userIDs, err
	}

	thirdUserIDs := RecursiveFindOpenIDs(departmentLists.DepartmentList)
	for _, user := range departmentLists.UserList {
		thirdUserIDs = append(thirdUserIDs, user.OpenID)
	}

	return public.ContainsSliceDiff(thirdUserIDs, userIDs), nil
}
