package packer

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/api/v1alpha1"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/consts"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/public"
	"github.com/go-omnibus/proof"
	"go.mongodb.org/mongo-driver/bson"
	"net/url"
)

func TransExpectToMaoMockCase(expects []*rao.Expect) []*v1alpha1.MockAPI_Case {
	cases := make([]*v1alpha1.MockAPI_Case, 0, len(expects))
	for _, e := range expects {
		items := make([]*v1alpha1.MockAPI_Condition_SimpleCondition_Item, 0)
		for _, ec := range e.Conditions {
			operandX := consts.MockRequestScope + consts.MockSplit + ec.Path + consts.MockSplit + ec.ParameterName
			if ec.Path == consts.MockConditionJsonPath {
				operandX = consts.MockRequestJsonScope + consts.MockSplit + ec.Path + consts.MockSplit + ec.ParameterName
			}
			item := &v1alpha1.MockAPI_Condition_SimpleCondition_Item{
				OperandX: operandX,
				Operator: ec.Compare,
				OperandY: ec.ParameterValue,
				Opposite: false,
			}
			items = append(items, item)
		}

		mockResponse := e.Response.Json
		contentType := "application/json"
		if e.Response.ContentType != consts.MockContentTypeJson {
			mockResponse = e.Response.Raw
		}
		if s, ok := consts.MockContentType[e.Response.ContentType]; ok {
			contentType = s
		}
		responseHeader := map[string]string{
			"content-type": contentType,
		}
		apiCase := &v1alpha1.MockAPI_Case{
			Condition: &v1alpha1.MockAPI_Condition{
				Condition: &v1alpha1.MockAPI_Condition_Simple{
					Simple: &v1alpha1.MockAPI_Condition_SimpleCondition{
						Items:           items,
						UseOrAmongItems: false,
					},
				}},
			Response: &v1alpha1.MockAPI_Response{
				Response: &v1alpha1.MockAPI_Response_Simple{
					Simple: &v1alpha1.MockAPI_Response_SimpleResponse{
						Code:   200,
						Header: responseHeader,
						Body:   mockResponse,
					},
				}},
		}
		cases = append(cases, apiCase)
	}

	return cases
}

func TransSaveMockTargetReqToMaoMock(target *rao.MockSaveTargetReq, mockAPICase []*v1alpha1.MockAPI_Case) *mao.Mock {
	reqRes := public.CheckStructIsEmpty(target.Request)
	if reqRes {
		log.Logger.Info("target.request not found request")
		return nil
	}

	expects, err := bson.Marshal(mao.Expects{Expects: target.Expects})
	if err != nil {
		log.Logger.Info("target.expects bson marshal err", proof.WithError(err))
	}

	u, err := url.Parse(target.Request.URL)
	if err != nil {
		log.Logger.Info("target.request.url parse err", proof.WithError(err))

	}

	mockCase, err := bson.Marshal(mao.Cases{Cases: mockAPICase})
	if err != nil {
		log.Logger.Info("TransExpectToMaoMockCase bson marshal err", proof.WithError(err))
	}

	return &mao.Mock{
		TargetID:   target.TargetID,
		TeamID:     target.TeamID,
		UniqueKey:  target.TargetID,
		Path:       u.Path,
		Method:     target.Method,
		Cases:      mockCase,
		Expects:    expects,
		IsMockOpen: target.IsMockOpen,
		MockPath:   target.MockPath,
	}
}

func TransFolderToMockTargetModel(folder *rao.MockFolder, userID string) *model.MockTarget {
	return &model.MockTarget{
		TargetID:      folder.TargetID,
		TeamID:        folder.TeamID,
		TargetType:    consts.TargetTypeFolder,
		Name:          folder.Name,
		ParentID:      folder.ParentID,
		Method:        folder.Method,
		Sort:          folder.Sort,
		TypeSort:      folder.TypeSort,
		Status:        consts.TargetStatusNormal,
		Version:       folder.Version,
		CreatedUserID: userID,
		RecentUserID:  userID,
		Source:        folder.Source,
		Description:   folder.Description,
	}
}

func TransSaveFolderReqToMockTargetModel(folder *rao.MockSaveFolderReq, userID string) *model.MockTarget {
	return &model.MockTarget{
		TargetID:      folder.TargetID,
		TeamID:        folder.TeamID,
		TargetType:    consts.TargetTypeFolder,
		Name:          folder.Name,
		ParentID:      folder.ParentID,
		Method:        folder.Method,
		Sort:          folder.Sort,
		TypeSort:      folder.TypeSort,
		Status:        consts.TargetStatusNormal,
		Version:       folder.Version,
		CreatedUserID: userID,
		RecentUserID:  userID,
		Source:        folder.Source,
		Description:   folder.Description,
	}
}

func TransTargetToRaoMockFolder(t *model.MockTarget) *rao.MockFolder {
	return &rao.MockFolder{
		TargetID:    t.TargetID,
		TeamID:      t.TeamID,
		ParentID:    t.ParentID,
		Name:        t.Name,
		Method:      t.Method,
		Sort:        t.Sort,
		TypeSort:    t.TypeSort,
		Version:     t.Version,
		Description: t.Description,
	}
}

func TransSaveTargetReqToMockTargetModel(target *rao.MockSaveTargetReq, userID string) *model.MockTarget {
	return &model.MockTarget{
		TargetID:      target.TargetID,
		TeamID:        target.TeamID,
		TargetType:    target.TargetType,
		Name:          target.Name,
		ParentID:      target.ParentID,
		Method:        target.Method,
		Sort:          target.Sort,
		TypeSort:      target.TypeSort,
		Status:        consts.TargetStatusNormal,
		Version:       target.Version,
		CreatedUserID: userID,
		RecentUserID:  userID,
		Source:        target.Source,
		Description:   target.Description,
	}
}

func TransSaveMockTargetReqToMaoAPI(target *rao.MockSaveTargetReq) *mao.API {
	reqRes := public.CheckStructIsEmpty(target.Request)
	if reqRes {
		log.Logger.Info("target.request not found request")
		return nil
	}
	header, err := bson.Marshal(target.Request.Header)
	if err != nil {
		log.Logger.Info("target.request.header bson marshal err", proof.WithError(err))
	}

	query, err := bson.Marshal(target.Request.Query)
	if err != nil {
		log.Logger.Info("target.request.query bson marshal err", proof.WithError(err))
	}

	cookie, err := bson.Marshal(target.Request.Cookie)
	if err != nil {
		log.Logger.Info("target.request.cookie bson marshal err", proof.WithError(err))
	}

	body, err := bson.Marshal(target.Request.Body)
	if err != nil {
		log.Logger.Info("target.request.body bson marshal err", proof.WithError(err))
	}

	auth, err := bson.Marshal(target.Request.Auth)
	if err != nil {
		log.Logger.Info("target.request.auth bson marshal err", proof.WithError(err))
	}

	assert, err := bson.Marshal(mao.Assert{Assert: target.Request.Assert})
	if err != nil {
		log.Logger.Info("target.request.assert bson marshal err", proof.WithError(err))
	}

	regex, err := bson.Marshal(mao.Regex{Regex: target.Request.Regex})
	if err != nil {
		log.Logger.Info("target.request.regex bson marshal err", proof.WithError(err))
	}

	return &mao.API{
		TargetID: target.TargetID,
		URL:      target.Request.URL,
		EnvInfo: mao.EnvInfo{
			EnvID:       target.EnvInfo.EnvID,
			EnvName:     target.EnvInfo.EnvName,
			ServiceID:   target.EnvInfo.ServiceID,
			ServiceName: target.EnvInfo.ServiceName,
			PreUrl:      target.EnvInfo.PreUrl,
		},
		Header:      header,
		Query:       query,
		Cookie:      cookie,
		Body:        body,
		Auth:        auth,
		Description: target.Description,
		Assert:      assert,
		Regex:       regex,
		HttpApiSetup: mao.HttpApiSetup{
			IsRedirects:         target.Request.HttpApiSetup.IsRedirects,
			RedirectsNum:        target.Request.HttpApiSetup.RedirectsNum,
			ReadTimeOut:         target.Request.HttpApiSetup.ReadTimeOut,
			WriteTimeOut:        target.Request.HttpApiSetup.WriteTimeOut,
			ClientName:          target.Request.HttpApiSetup.ClientName,
			KeepAlive:           target.Request.HttpApiSetup.KeepAlive,
			MaxIdleConnDuration: target.Request.HttpApiSetup.MaxIdleConnDuration,
			MaxConnPerHost:      target.Request.HttpApiSetup.MaxConnPerHost,
			UserAgent:           target.Request.HttpApiSetup.UserAgent,
			MaxConnWaitTimeout:  target.Request.HttpApiSetup.MaxConnWaitTimeout,
		},
	}
}

func TransMockTargetsToRaoAPIDetails(targets []*model.MockTarget, apiMap map[string]*mao.API, mockMap map[string]*mao.Mock) []rao.MockAPIDetail {
	ret := make([]rao.MockAPIDetail, 0, len(targets))

	globalVariable := rao.GlobalVariable{}

	for _, target := range targets {
		if api, ok := apiMap[target.TargetID]; ok {
			if mock, ok := mockMap[target.TargetID]; ok {
				ret = append(ret, TransMockTargetToRaoMockAPIDetail(target, api, mock, globalVariable))
			}
		}
	}
	return ret
}

func TransMockTargetToRaoMockAPIDetail(target *model.MockTarget, api *mao.API, mock *mao.Mock, globalVariable rao.GlobalVariable) rao.MockAPIDetail {
	auth := rao.Auth{}
	if err := bson.Unmarshal(api.Auth, &auth); err != nil {
		log.Logger.Info("api.auth bson Unmarshal err", proof.WithError(err))
	}
	body := rao.Body{}
	if err := bson.Unmarshal(api.Body, &body); err != nil {
		log.Logger.Info("api.body bson Unmarshal err", proof.WithError(err))
	}
	header := rao.Header{}
	if err := bson.Unmarshal(api.Header, &header); err != nil {
		log.Logger.Info("api.header bson Unmarshal err", proof.WithError(err))
	}
	query := rao.Query{}
	if err := bson.Unmarshal(api.Query, &query); err != nil {
		log.Logger.Info("api.query bson Unmarshal err", proof.WithError(err))
	}

	cookie := rao.Cookie{}
	if err := bson.Unmarshal(api.Cookie, &cookie); err != nil {
		log.Logger.Info("api.cookie bson Unmarshal err", proof.WithError(err))
	}

	assert := mao.Assert{}
	if err := bson.Unmarshal(api.Assert, &assert); err != nil {
		log.Logger.Info("api.assert bson Unmarshal err", proof.WithError(err))
	}

	regex := mao.Regex{}
	if err := bson.Unmarshal(api.Regex, &regex); err != nil {
		log.Logger.Info("api.regex bson Unmarshal err", proof.WithError(err))
	}

	expects := mao.Expects{}
	if err := bson.Unmarshal(mock.Expects, &expects); err != nil {
		log.Logger.Info("mock.expects bson Unmarshal err", proof.WithError(err))
	}

	apiDetail := rao.APIDetail{
		TargetID:   target.TargetID,
		ParentID:   target.ParentID,
		TeamID:     target.TeamID,
		TargetType: target.TargetType,
		Name:       target.Name,
		Method:     target.Method,
		URL:        api.URL,
		EnvInfo: rao.EnvInfo{
			EnvID:       api.EnvInfo.EnvID,
			EnvName:     api.EnvInfo.EnvName,
			ServiceID:   api.EnvInfo.ServiceID,
			ServiceName: api.EnvInfo.ServiceName,
			PreUrl:      api.EnvInfo.PreUrl,
		},
		Sort:     target.Sort,
		TypeSort: target.TypeSort,
		Request: rao.Request{
			PreUrl:      api.EnvInfo.PreUrl,
			URL:         api.URL,
			Description: api.Description,
			Auth:        auth,
			Body:        body,
			Header:      header,
			Query:       query,
			Cookie:      cookie,
			Assert:      assert.Assert,
			Regex:       regex.Regex,
			HttpApiSetup: rao.HttpApiSetup{
				IsRedirects:         api.HttpApiSetup.IsRedirects,
				RedirectsNum:        api.HttpApiSetup.RedirectsNum,
				ReadTimeOut:         api.HttpApiSetup.ReadTimeOut,
				WriteTimeOut:        api.HttpApiSetup.WriteTimeOut,
				ClientName:          api.HttpApiSetup.ClientName,
				KeepAlive:           api.HttpApiSetup.KeepAlive,
				MaxIdleConnDuration: api.HttpApiSetup.MaxIdleConnDuration,
				MaxConnPerHost:      api.HttpApiSetup.MaxConnPerHost,
				UserAgent:           api.HttpApiSetup.UserAgent,
				MaxConnWaitTimeout:  api.HttpApiSetup.MaxConnWaitTimeout,
			},
		},
		Version:        target.Version,
		Description:    api.Description,
		CreatedTimeSec: target.CreatedAt.Unix(),
		UpdatedTimeSec: target.UpdatedAt.Unix(),
		GlobalVariable: globalVariable,
	}

	return rao.MockAPIDetail{
		APIDetail:  &apiDetail,
		IsMockOpen: mock.IsMockOpen,
		MockPath:   mock.MockPath,
		Expects:    expects.Expects,
	}
}

func TransMockTargetToRaoFolderAPIList(targets []*model.MockTarget, mockMap map[string]*mao.Mock) []*rao.MockFolderAPI {
	ret := make([]*rao.MockFolderAPI, 0)
	for _, t := range targets {
		mockFolderAPI := &rao.MockFolderAPI{
			TargetID:      t.TargetID,
			TeamID:        t.TeamID,
			TargetType:    t.TargetType,
			Name:          t.Name,
			ParentID:      t.ParentID,
			Method:        t.Method,
			Sort:          t.Sort,
			TypeSort:      t.TypeSort,
			Version:       t.Version,
			Source:        t.Source,
			CreatedUserID: t.CreatedUserID,
			RecentUserID:  t.RecentUserID,
		}
		if m, ok := mockMap[t.TargetID]; ok {
			mockFolderAPI.IsMockOpen = m.IsMockOpen
		}

		ret = append(ret, mockFolderAPI)
	}
	return ret
}

func GetRunMockTargetParam(target *model.MockTarget, globalVariable rao.GlobalVariable, api *mao.API) rao.RunTargetParam {
	auth := rao.Auth{}
	if err := bson.Unmarshal(api.Auth, &auth); err != nil {
		log.Logger.Info("api.auth bson Unmarshal err", proof.WithError(err))
	}
	body := rao.Body{}
	if err := bson.Unmarshal(api.Body, &body); err != nil {
		log.Logger.Info("api.body bson Unmarshal err", proof.WithError(err))
	}
	header := rao.Header{}
	if err := bson.Unmarshal(api.Header, &header); err != nil {
		log.Logger.Info("api.header bson Unmarshal err", proof.WithError(err))
	}
	query := rao.Query{}
	if err := bson.Unmarshal(api.Query, &query); err != nil {
		log.Logger.Info("api.query bson Unmarshal err", proof.WithError(err))
	}

	cookie := rao.Cookie{}
	if err := bson.Unmarshal(api.Cookie, &cookie); err != nil {
		log.Logger.Info("api.cookie bson Unmarshal err", proof.WithError(err))
	}

	assert := mao.Assert{}
	if err := bson.Unmarshal(api.Assert, &assert); err != nil {
		log.Logger.Info("api.assert bson Unmarshal err", proof.WithError(err))
	}

	regex := mao.Regex{}
	if err := bson.Unmarshal(api.Regex, &regex); err != nil {
		log.Logger.Info("api.regex bson Unmarshal err", proof.WithError(err))
	}

	preUrl := ""
	if api.EnvInfo.EnvID != 0 {
		preUrl = api.EnvInfo.PreUrl
	}

	return rao.RunTargetParam{
		TargetID:   target.TargetID,
		ParentID:   target.ParentID,
		TeamID:     target.TeamID,
		TargetType: target.TargetType,
		Name:       target.Name,
		Request: rao.Request{
			PreUrl:      preUrl,
			URL:         api.URL,
			Method:      target.Method,
			Description: api.Description,
			Auth:        auth,
			Body:        body,
			Header:      header,
			Query:       query,
			Cookie:      cookie,
			Assert:      assert.Assert,
			Regex:       regex.Regex,
			HttpApiSetup: rao.HttpApiSetup{
				IsRedirects:         api.HttpApiSetup.IsRedirects,
				RedirectsNum:        api.HttpApiSetup.RedirectsNum,
				ReadTimeOut:         api.HttpApiSetup.ReadTimeOut,
				WriteTimeOut:        api.HttpApiSetup.WriteTimeOut,
				ClientName:          api.HttpApiSetup.ClientName,
				KeepAlive:           api.HttpApiSetup.KeepAlive,
				MaxIdleConnDuration: api.HttpApiSetup.MaxIdleConnDuration,
				MaxConnPerHost:      api.HttpApiSetup.MaxConnPerHost,
				UserAgent:           api.HttpApiSetup.UserAgent,
				MaxConnWaitTimeout:  api.HttpApiSetup.MaxConnWaitTimeout,
			},
		},
		GlobalVariable: globalVariable,
	}
}
