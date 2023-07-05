package packer

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/public"
	"github.com/go-omnibus/proof"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func TransSaveTargetReqToMaoAPI(target *rao.SaveTargetReq) *mao.API {
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

func TransTargetToRaoAPIDetail(target *model.Target, api *mao.API, globalVariable rao.GlobalVariable) rao.APIDetail {
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

	return rao.APIDetail{
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
		Version:        target.Version,
		Description:    api.Description,
		CreatedTimeSec: target.CreatedAt.Unix(),
		UpdatedTimeSec: target.UpdatedAt.Unix(),
		GlobalVariable: globalVariable,
	}
}

func TransTargetToRaoSqlDetail(target *model.Target, sql *mao.SqlDetailForMg, globalVariable rao.GlobalVariable) rao.APIDetail {
	// 断言
	assert := make([]rao.SqlAssert, 0, 10)
	for _, assertInfo := range sql.Assert {
		temp := rao.SqlAssert{
			IsChecked: assertInfo.IsChecked,
			Field:     assertInfo.Field,
			Compare:   assertInfo.Compare,
			Val:       assertInfo.Val,
			Index:     assertInfo.Index,
		}
		assert = append(assert, temp)
	}

	// 关联提取
	regex := make([]rao.SqlRegex, 0, 10)
	for _, regexInfo := range sql.Regex {
		temp := rao.SqlRegex{
			IsChecked: regexInfo.IsChecked,
			Var:       regexInfo.Var,
			Field:     regexInfo.Field,
			Index:     regexInfo.Index,
		}
		regex = append(regex, temp)
	}

	return rao.APIDetail{
		TargetID:       target.TargetID,
		ParentID:       target.ParentID,
		TeamID:         target.TeamID,
		TargetType:     target.TargetType,
		Name:           target.Name,
		Method:         target.Method,
		Sort:           target.Sort,
		TypeSort:       target.TypeSort,
		Version:        target.Version,
		CreatedTimeSec: target.CreatedAt.Unix(),
		UpdatedTimeSec: target.UpdatedAt.Unix(),
		GlobalVariable: globalVariable,
		EnvInfo: rao.EnvInfo{
			EnvID:       sql.EnvInfo.EnvID,
			EnvName:     sql.EnvInfo.EnvName,
			ServiceID:   sql.EnvInfo.ServiceID,
			ServiceName: sql.EnvInfo.ServiceName,
			PreUrl:      sql.EnvInfo.PreUrl,
			DatabaseID:  sql.EnvInfo.DatabaseID,
			ServerName:  sql.EnvInfo.ServerName,
		},
		SqlDetail: rao.SqlDetail{
			SqlString: sql.SqlString,
			Assert:    assert,
			Regex:     regex,
			SqlDatabaseInfo: rao.SqlDatabaseInfo{
				Type:       sql.SqlDatabaseInfo.Type,
				ServerName: sql.SqlDatabaseInfo.ServerName,
				Host:       sql.SqlDatabaseInfo.Host,
				User:       sql.SqlDatabaseInfo.User,
				Password:   sql.SqlDatabaseInfo.Password,
				Port:       sql.SqlDatabaseInfo.Port,
				DbName:     sql.SqlDatabaseInfo.DbName,
				Charset:    sql.SqlDatabaseInfo.Charset,
			},
		},
	}
}

func TransTargetToRaoTcpDetail(target *model.Target, tcp *mao.TcpDetail, globalVariable rao.GlobalVariable) rao.APIDetail {
	return rao.APIDetail{
		TargetID:       target.TargetID,
		ParentID:       target.ParentID,
		TeamID:         target.TeamID,
		TargetType:     target.TargetType,
		Name:           target.Name,
		Method:         target.Method,
		Sort:           target.Sort,
		TypeSort:       target.TypeSort,
		Version:        target.Version,
		CreatedTimeSec: target.CreatedAt.Unix(),
		UpdatedTimeSec: target.UpdatedAt.Unix(),
		GlobalVariable: globalVariable,
		EnvInfo: rao.EnvInfo{
			EnvID:       tcp.EnvInfo.EnvID,
			EnvName:     tcp.EnvInfo.EnvName,
			ServiceID:   tcp.EnvInfo.ServiceID,
			ServiceName: tcp.EnvInfo.ServiceName,
			PreUrl:      tcp.EnvInfo.PreUrl,
		},
		TcpDetail: rao.TcpDetail{
			Url:         tcp.Url,
			MessageType: tcp.MessageType,
			SendMessage: tcp.SendMessage,
			TcpConfig: rao.TcpConfig{
				ConnectType:         tcp.TcpConfig.ConnectType,
				IsAutoSend:          tcp.TcpConfig.IsAutoSend,
				ConnectDurationTime: tcp.TcpConfig.ConnectDurationTime,
				SendMsgDurationTime: tcp.TcpConfig.SendMsgDurationTime,
				ConnectTimeoutTime:  tcp.TcpConfig.ConnectTimeoutTime,
				RetryNum:            tcp.TcpConfig.RetryNum,
				RetryInterval:       tcp.TcpConfig.RetryInterval,
			},
		},
	}
}

func TransTargetToRaoWebsocketDetail(target *model.Target, websocket *mao.WebsocketDetail, globalVariable rao.GlobalVariable) rao.APIDetail {
	wsHeader := make([]rao.WsQuery, 0, 10)
	for _, v := range websocket.WsHeader {
		temp := rao.WsQuery{
			IsChecked: v.IsChecked,
			Var:       v.Var,
			Val:       v.Val,
		}
		wsHeader = append(wsHeader, temp)
	}

	wsParam := make([]rao.WsQuery, 0, 10)
	for _, v := range websocket.WsParam {
		temp := rao.WsQuery{
			IsChecked: v.IsChecked,
			Var:       v.Var,
			Val:       v.Val,
		}
		wsParam = append(wsParam, temp)
	}

	wsEvent := make([]rao.WsQuery, 0, 10)
	for _, v := range websocket.WsEvent {
		temp := rao.WsQuery{
			IsChecked: v.IsChecked,
			Var:       v.Var,
			Val:       v.Val,
		}
		wsEvent = append(wsEvent, temp)
	}

	return rao.APIDetail{
		TargetID:       target.TargetID,
		ParentID:       target.ParentID,
		TeamID:         target.TeamID,
		TargetType:     target.TargetType,
		Name:           target.Name,
		Method:         target.Method,
		Sort:           target.Sort,
		TypeSort:       target.TypeSort,
		Version:        target.Version,
		CreatedTimeSec: target.CreatedAt.Unix(),
		UpdatedTimeSec: target.UpdatedAt.Unix(),
		GlobalVariable: globalVariable,
		EnvInfo: rao.EnvInfo{
			EnvID:       websocket.EnvInfo.EnvID,
			EnvName:     websocket.EnvInfo.EnvName,
			ServiceID:   websocket.EnvInfo.ServiceID,
			ServiceName: websocket.EnvInfo.ServiceName,
			PreUrl:      websocket.EnvInfo.PreUrl,
		},
		WebsocketDetail: rao.WebsocketDetail{
			Url:         websocket.Url,
			SendMessage: websocket.SendMessage,
			MessageType: websocket.MessageType,
			WsHeader:    wsHeader,
			WsParam:     wsParam,
			WsEvent:     wsEvent,
			WsConfig: rao.WsConfig{
				ConnectType:         websocket.WsConfig.ConnectType,
				IsAutoSend:          websocket.WsConfig.IsAutoSend,
				ConnectDurationTime: websocket.WsConfig.ConnectDurationTime,
				SendMsgDurationTime: websocket.WsConfig.SendMsgDurationTime,
				ConnectTimeoutTime:  websocket.WsConfig.ConnectTimeoutTime,
				RetryNum:            websocket.WsConfig.RetryNum,
				RetryInterval:       websocket.WsConfig.RetryInterval,
			},
		},
	}
}

func TransTargetsToRaoAPIDetails(targets []*model.Target, apiMap map[string]*mao.API,
	sqlMap map[string]*mao.SqlDetailForMg, tcpMap map[string]*mao.TcpDetail,
	websocketMap map[string]*mao.WebsocketDetail, dubboMap map[string]*mao.DubboDetail) []rao.APIDetail {
	ret := make([]rao.APIDetail, 0, len(targets))

	globalVariable := rao.GlobalVariable{}

	for _, target := range targets {
		if api, ok := apiMap[target.TargetID]; ok {
			ret = append(ret, TransTargetToRaoAPIDetail(target, api, globalVariable))
		}
		if sql, ok := sqlMap[target.TargetID]; ok {
			ret = append(ret, TransTargetToRaoSqlDetail(target, sql, globalVariable))
		}

		if tcp, ok := tcpMap[target.TargetID]; ok {
			ret = append(ret, TransTargetToRaoTcpDetail(target, tcp, globalVariable))
		}
		if websocket, ok := websocketMap[target.TargetID]; ok {
			ret = append(ret, TransTargetToRaoWebsocketDetail(target, websocket, globalVariable))
		}
		//if mqtt, ok := mqttMap[target.TargetID]; ok {
		//	ret = append(ret, TransTargetToRaoMqttDetail(target, mqtt, globalVariable))
		//}
		if dubbo, ok := dubboMap[target.TargetID]; ok {
			ret = append(ret, TransTargetToRaoDubboDetail(target, dubbo, globalVariable))
		}
	}
	return ret
}

func TransSaveTargetReqToMaoSqlDetail(target *rao.SaveTargetReq) *mao.SqlDetailForMg {
	// 断言
	assert := make([]mao.SqlAssert, 0, 10)
	for _, assertInfo := range target.SqlDetail.Assert {
		temp := mao.SqlAssert{
			IsChecked: assertInfo.IsChecked,
			Field:     assertInfo.Field,
			Compare:   assertInfo.Compare,
			Val:       assertInfo.Val,
			Index:     assertInfo.Index,
		}
		assert = append(assert, temp)
	}

	// 关联提取
	regex := make([]mao.SqlRegex, 0, 10)
	for _, regexInfo := range target.SqlDetail.Regex {
		temp := mao.SqlRegex{
			IsChecked: regexInfo.IsChecked,
			Var:       regexInfo.Var,
			Field:     regexInfo.Field,
			Index:     regexInfo.Index,
		}
		regex = append(regex, temp)
	}

	dbType := "mysql"
	if target.Method == "oracle" {
		dbType = "oracle"
	} else if target.Method == "PgSQL" {
		dbType = "postgresql"
	}

	return &mao.SqlDetailForMg{
		TargetID: target.TargetID,
		TeamID:   target.TeamID,
		Assert:   assert,
		Regex:    regex,
		EnvInfo: mao.EnvInfo{
			EnvID:       target.EnvInfo.EnvID,
			EnvName:     target.EnvInfo.EnvName,
			ServiceID:   target.EnvInfo.ServiceID,
			ServiceName: target.EnvInfo.ServiceName,
			PreUrl:      target.EnvInfo.PreUrl,
			DatabaseID:  target.EnvInfo.DatabaseID,
			ServerName:  target.EnvInfo.ServerName,
		},
		SqlString: target.SqlDetail.SqlString,
		SqlDatabaseInfo: mao.SqlDatabaseInfo{
			Type:       dbType,
			ServerName: target.SqlDetail.SqlDatabaseInfo.ServerName,
			Host:       target.SqlDetail.SqlDatabaseInfo.Host,
			User:       target.SqlDetail.SqlDatabaseInfo.User,
			Password:   target.SqlDetail.SqlDatabaseInfo.Password,
			Port:       target.SqlDetail.SqlDatabaseInfo.Port,
			DbName:     target.SqlDetail.SqlDatabaseInfo.DbName,
			Charset:    target.SqlDetail.SqlDatabaseInfo.Charset,
		},
		CreatedAt: time.Now(),
	}
}

func TransRunSqlParam(target *model.Target, sqlDetailInfo *mao.SqlDetailForMg, globalVariable rao.GlobalVariable) rao.RunTargetParam {
	// 断言
	assert := make([]rao.SqlAssert, 0, 10)
	for _, assertInfo := range sqlDetailInfo.Assert {
		temp := rao.SqlAssert{
			IsChecked: assertInfo.IsChecked,
			Field:     assertInfo.Field,
			Compare:   assertInfo.Compare,
			Val:       assertInfo.Val,
			Index:     assertInfo.Index,
		}
		assert = append(assert, temp)
	}

	// 关联提取
	regex := make([]rao.SqlRegex, 0, 10)
	for _, regexInfo := range sqlDetailInfo.Regex {
		temp := rao.SqlRegex{
			IsChecked: regexInfo.IsChecked,
			Var:       regexInfo.Var,
			Field:     regexInfo.Field,
			Index:     regexInfo.Index,
		}
		regex = append(regex, temp)
	}

	dbType := "mysql"
	if target.Method == "ORACLE" {
		dbType = "oracle"
	} else if target.Method == "PgSQL" {
		dbType = "postgresql"
	}

	return rao.RunTargetParam{
		TargetID:   target.TargetID,
		Name:       target.Name,
		TeamID:     target.TeamID,
		TargetType: target.TargetType,
		SqlDetail: rao.SqlDetail{
			SqlString: sqlDetailInfo.SqlString,
			SqlDatabaseInfo: rao.SqlDatabaseInfo{
				Type:     dbType,
				Host:     sqlDetailInfo.SqlDatabaseInfo.Host,
				User:     sqlDetailInfo.SqlDatabaseInfo.User,
				Password: sqlDetailInfo.SqlDatabaseInfo.Password,
				Port:     sqlDetailInfo.SqlDatabaseInfo.Port,
				DbName:   sqlDetailInfo.SqlDatabaseInfo.DbName,
				Charset:  sqlDetailInfo.SqlDatabaseInfo.Charset,
			},
			Assert: assert,
			Regex:  regex,
		},
		GlobalVariable: globalVariable,
	}
}

func TransSaveTargetReqToMaoTcpDetail(target *rao.SaveTargetReq) *mao.TcpDetail {
	return &mao.TcpDetail{
		TargetID: target.TargetID,
		TeamID:   target.TeamID,
		EnvInfo: mao.EnvInfo{
			EnvID:       target.EnvInfo.EnvID,
			EnvName:     target.EnvInfo.EnvName,
			ServiceID:   target.EnvInfo.ServiceID,
			ServiceName: target.EnvInfo.ServiceName,
			PreUrl:      target.EnvInfo.PreUrl,
		},
		Url:         target.TcpDetail.Url,
		MessageType: target.TcpDetail.MessageType,
		SendMessage: target.TcpDetail.SendMessage,
		TcpConfig: mao.TcpConfig{
			ConnectType:         target.TcpDetail.TcpConfig.ConnectType,
			IsAutoSend:          target.TcpDetail.TcpConfig.IsAutoSend,
			ConnectDurationTime: target.TcpDetail.TcpConfig.ConnectDurationTime,
			SendMsgDurationTime: target.TcpDetail.TcpConfig.SendMsgDurationTime,
			ConnectTimeoutTime:  target.TcpDetail.TcpConfig.ConnectTimeoutTime,
			RetryNum:            target.TcpDetail.TcpConfig.RetryNum,
			RetryInterval:       target.TcpDetail.TcpConfig.RetryInterval,
		},
		CreatedAt: time.Now(),
	}
}

func TransSaveTargetReqToMaoWebsocketDetail(target *rao.SaveTargetReq) *mao.WebsocketDetail {
	wsHeader := make([]mao.WsQuery, 0, 10)
	for _, v := range target.WebsocketDetail.WsHeader {
		temp := mao.WsQuery{
			IsChecked: v.IsChecked,
			Var:       v.Var,
			Val:       v.Val,
		}
		wsHeader = append(wsHeader, temp)
	}

	wsParam := make([]mao.WsQuery, 0, 10)
	for _, v := range target.WebsocketDetail.WsParam {
		temp := mao.WsQuery{
			IsChecked: v.IsChecked,
			Var:       v.Var,
			Val:       v.Val,
		}
		wsParam = append(wsParam, temp)
	}

	wsEvent := make([]mao.WsQuery, 0, 10)
	for _, v := range target.WebsocketDetail.WsEvent {
		temp := mao.WsQuery{
			IsChecked: v.IsChecked,
			Var:       v.Var,
			Val:       v.Val,
		}
		wsEvent = append(wsEvent, temp)
	}

	return &mao.WebsocketDetail{
		TargetID: target.TargetID,
		TeamID:   target.TeamID,
		EnvInfo: mao.EnvInfo{
			EnvID:       target.EnvInfo.EnvID,
			EnvName:     target.EnvInfo.EnvName,
			ServiceID:   target.EnvInfo.ServiceID,
			ServiceName: target.EnvInfo.ServiceName,
			PreUrl:      target.EnvInfo.PreUrl,
		},
		Url:         target.WebsocketDetail.Url,
		SendMessage: target.WebsocketDetail.SendMessage,
		MessageType: target.WebsocketDetail.MessageType,
		WsHeader:    wsHeader,
		WsParam:     wsParam,
		WsEvent:     wsEvent,
		WsConfig: mao.WsConfig{
			ConnectType:         target.WebsocketDetail.WsConfig.ConnectType,
			IsAutoSend:          target.WebsocketDetail.WsConfig.IsAutoSend,
			ConnectDurationTime: target.WebsocketDetail.WsConfig.ConnectDurationTime,
			SendMsgDurationTime: target.WebsocketDetail.WsConfig.SendMsgDurationTime,
			ConnectTimeoutTime:  target.WebsocketDetail.WsConfig.ConnectTimeoutTime,
			RetryNum:            target.WebsocketDetail.WsConfig.RetryNum,
			RetryInterval:       target.WebsocketDetail.WsConfig.RetryInterval,
		},
		CreatedAt: time.Now(),
	}
}

func TransSaveTargetReqToMaoMqttDetail(target *rao.SaveTargetReq) *mao.MqttDetail {
	return &mao.MqttDetail{
		TargetID: target.TargetID,
		TeamID:   target.TeamID,
		EnvInfo: mao.EnvInfo{
			EnvID:       target.EnvInfo.EnvID,
			EnvName:     target.EnvInfo.EnvName,
			ServiceID:   target.EnvInfo.ServiceID,
			ServiceName: target.EnvInfo.ServiceName,
			PreUrl:      target.EnvInfo.PreUrl,
		},
		Topic:       target.MqttDetail.Topic,
		SendMessage: target.MqttDetail.SendMessage,
		CommonConfig: mao.CommonConfig{
			ClientName: target.MqttDetail.CommonConfig.ClientName,
			Username:   target.MqttDetail.CommonConfig.Username,
			Password:   target.MqttDetail.CommonConfig.Password,
			IsEncrypt:  target.MqttDetail.CommonConfig.IsEncrypt,
			AuthFile: mao.AuthFile{
				FileName: target.MqttDetail.CommonConfig.AuthFile.FileName,
				FileUrl:  target.MqttDetail.CommonConfig.AuthFile.FileUrl,
			},
		},
		HigherConfig: mao.HigherConfig{
			ConnectTimeoutTime:  target.MqttDetail.HigherConfig.ConnectTimeoutTime,
			KeepAliveTime:       target.MqttDetail.HigherConfig.KeepAliveTime,
			IsAutoRetry:         target.MqttDetail.HigherConfig.IsAutoRetry,
			RetryNum:            target.MqttDetail.HigherConfig.RetryNum,
			RetryInterval:       target.MqttDetail.HigherConfig.RetryInterval,
			MqttVersion:         target.MqttDetail.HigherConfig.MqttVersion,
			DialogueTimeout:     target.MqttDetail.HigherConfig.DialogueTimeout,
			IsSaveMessage:       target.MqttDetail.HigherConfig.IsSaveMessage,
			ServiceQuality:      target.MqttDetail.HigherConfig.ServiceQuality,
			SendMsgIntervalTime: target.MqttDetail.HigherConfig.SendMsgIntervalTime,
		},
		WillConfig: mao.WillConfig{
			WillTopic:      target.MqttDetail.WillConfig.WillTopic,
			IsOpenWill:     target.MqttDetail.WillConfig.IsOpenWill,
			ServiceQuality: target.MqttDetail.WillConfig.ServiceQuality,
		},
		CreatedAt: time.Now(),
	}
}

func TransSaveTargetReqToMaoDubboDetail(target *rao.SaveTargetReq) *mao.DubboDetail {
	dubboParam := make([]mao.DubboParam, 0, 10)
	for _, paramInfo := range target.DubboDetail.DubboParam {
		temp := mao.DubboParam{
			IsChecked: paramInfo.IsChecked,
			ParamType: paramInfo.ParamType,
			Var:       paramInfo.Var,
			Val:       paramInfo.Val,
		}
		dubboParam = append(dubboParam, temp)
	}

	dubboAssert := make([]mao.DubboAssert, 0, 10)
	for _, v := range target.DubboDetail.DubboAssert {
		temp := mao.DubboAssert{
			IsChecked:    v.IsChecked,
			ResponseType: v.ResponseType,
			Var:          v.Var,
			Compare:      v.Compare,
			Val:          v.Val,
		}
		dubboAssert = append(dubboAssert, temp)
	}

	dubboRegex := make([]mao.DubboRegex, 0, 10)
	for _, v := range target.DubboDetail.DubboRegex {
		temp := mao.DubboRegex{
			IsChecked: v.IsChecked,
			Type:      v.Type,
			Var:       v.Var,
			Express:   v.Express,
			Val:       v.Val,
			Index:     v.Index,
		}
		dubboRegex = append(dubboRegex, temp)
	}

	return &mao.DubboDetail{
		TargetID: target.TargetID,
		TeamID:   target.TeamID,
		EnvInfo: mao.EnvInfo{
			EnvID:       target.EnvInfo.EnvID,
			EnvName:     target.EnvInfo.EnvName,
			ServiceID:   target.EnvInfo.ServiceID,
			ServiceName: target.EnvInfo.ServiceName,
			PreUrl:      target.EnvInfo.PreUrl,
		},
		ApiName:       target.DubboDetail.ApiName,
		FunctionName:  target.DubboDetail.FunctionName,
		DubboProtocol: target.DubboDetail.DubboProtocol,
		DubboParam:    dubboParam,
		DubboAssert:   dubboAssert,
		DubboRegex:    dubboRegex,
		DubboConfig: mao.DubboConfig{
			RegistrationCenterName:    target.DubboDetail.DubboConfig.RegistrationCenterName,
			RegistrationCenterAddress: target.DubboDetail.DubboConfig.RegistrationCenterAddress,
			Version:                   target.DubboDetail.DubboConfig.Version,
		},
		CreatedAt: time.Now(),
	}
}

func TransTargetToRaoDubboDetail(target *model.Target, dubbo *mao.DubboDetail, globalVariable rao.GlobalVariable) rao.APIDetail {
	dubboParam := make([]rao.DubboParam, 0, 10)
	for _, v := range dubbo.DubboParam {
		temp := rao.DubboParam{
			IsChecked: v.IsChecked,
			ParamType: v.ParamType,
			Var:       v.Var,
			Val:       v.Val,
		}
		dubboParam = append(dubboParam, temp)
	}

	dubboAssert := make([]rao.DubboAssert, 0, 10)
	for _, v := range dubbo.DubboAssert {
		temp := rao.DubboAssert{
			IsChecked:    v.IsChecked,
			ResponseType: v.ResponseType,
			Var:          v.Var,
			Compare:      v.Compare,
			Val:          v.Val,
		}
		dubboAssert = append(dubboAssert, temp)
	}

	dubboRegex := make([]rao.DubboRegex, 0, 10)
	for _, v := range dubbo.DubboRegex {
		temp := rao.DubboRegex{
			IsChecked: v.IsChecked,
			Type:      v.Type,
			Var:       v.Var,
			Express:   v.Express,
			Val:       v.Val,
			Index:     v.Index,
		}
		dubboRegex = append(dubboRegex, temp)
	}

	return rao.APIDetail{
		TargetID:       target.TargetID,
		ParentID:       target.ParentID,
		TeamID:         target.TeamID,
		TargetType:     target.TargetType,
		Name:           target.Name,
		Method:         target.Method,
		Sort:           target.Sort,
		TypeSort:       target.TypeSort,
		Version:        target.Version,
		CreatedTimeSec: target.CreatedAt.Unix(),
		UpdatedTimeSec: target.UpdatedAt.Unix(),
		GlobalVariable: globalVariable,
		EnvInfo: rao.EnvInfo{
			EnvID:       dubbo.EnvInfo.EnvID,
			EnvName:     dubbo.EnvInfo.EnvName,
			ServiceID:   dubbo.EnvInfo.ServiceID,
			ServiceName: dubbo.EnvInfo.ServiceName,
			PreUrl:      dubbo.EnvInfo.PreUrl,
		},
		DubboDetail: rao.DubboDetail{
			ApiName:       dubbo.ApiName,
			FunctionName:  dubbo.FunctionName,
			DubboParam:    dubboParam,
			DubboProtocol: dubbo.DubboProtocol,
			DubboAssert:   dubboAssert,
			DubboRegex:    dubboRegex,
			DubboConfig: rao.DubboConfig{
				RegistrationCenterName:    dubbo.DubboConfig.RegistrationCenterName,
				RegistrationCenterAddress: dubbo.DubboConfig.RegistrationCenterAddress,
				Version:                   dubbo.DubboConfig.Version,
			},
		},
	}
}

func TransTargetToRaoMqttDetail(target *model.Target, mqtt *mao.MqttDetail, globalVariable rao.GlobalVariable) rao.APIDetail {
	return rao.APIDetail{
		TargetID:       target.TargetID,
		ParentID:       target.ParentID,
		TeamID:         target.TeamID,
		TargetType:     target.TargetType,
		Name:           target.Name,
		Method:         target.Method,
		Sort:           target.Sort,
		TypeSort:       target.TypeSort,
		Version:        target.Version,
		CreatedTimeSec: target.CreatedAt.Unix(),
		UpdatedTimeSec: target.UpdatedAt.Unix(),
		GlobalVariable: globalVariable,
		EnvInfo: rao.EnvInfo{
			EnvID:       mqtt.EnvInfo.EnvID,
			EnvName:     mqtt.EnvInfo.EnvName,
			ServiceID:   mqtt.EnvInfo.ServiceID,
			ServiceName: mqtt.EnvInfo.ServiceName,
			PreUrl:      mqtt.EnvInfo.PreUrl,
		},
		MqttDetail: rao.MqttDetail{
			Topic:       mqtt.Topic,
			SendMessage: mqtt.SendMessage,
			CommonConfig: rao.CommonConfig{
				ClientName: mqtt.CommonConfig.ClientName,
				Username:   mqtt.CommonConfig.Username,
				Password:   mqtt.CommonConfig.Password,
				IsEncrypt:  mqtt.CommonConfig.IsEncrypt,
				AuthFile: rao.AuthFile{
					FileName: mqtt.CommonConfig.AuthFile.FileName,
					FileUrl:  mqtt.CommonConfig.AuthFile.FileUrl,
				},
			},
			HigherConfig: rao.HigherConfig{
				ConnectTimeoutTime:  mqtt.HigherConfig.ConnectTimeoutTime,
				KeepAliveTime:       mqtt.HigherConfig.KeepAliveTime,
				IsAutoRetry:         mqtt.HigherConfig.IsAutoRetry,
				RetryNum:            mqtt.HigherConfig.RetryNum,
				RetryInterval:       mqtt.HigherConfig.RetryInterval,
				MqttVersion:         mqtt.HigherConfig.MqttVersion,
				DialogueTimeout:     mqtt.HigherConfig.DialogueTimeout,
				IsSaveMessage:       mqtt.HigherConfig.IsSaveMessage,
				ServiceQuality:      mqtt.HigherConfig.ServiceQuality,
				SendMsgIntervalTime: mqtt.HigherConfig.SendMsgIntervalTime,
			},
			WillConfig: rao.WillConfig{
				WillTopic:      mqtt.WillConfig.WillTopic,
				IsOpenWill:     mqtt.WillConfig.IsOpenWill,
				ServiceQuality: mqtt.WillConfig.ServiceQuality,
			},
		},
	}
}

func GetSendTcpParam(target *model.Target, tcpDetailInfo *mao.TcpDetail, globalVariable rao.GlobalVariable) rao.RunTargetParam {
	url := tcpDetailInfo.Url
	if tcpDetailInfo.EnvInfo.EnvID != 0 {
		url = tcpDetailInfo.EnvInfo.PreUrl + url
	}

	return rao.RunTargetParam{
		TargetID:   target.TargetID,
		Name:       target.Name,
		TeamID:     target.TeamID,
		TargetType: target.TargetType,
		TcpDetail: rao.TcpDetail{
			Url:         tcpDetailInfo.Url,
			MessageType: tcpDetailInfo.MessageType,
			SendMessage: tcpDetailInfo.SendMessage,
			TcpConfig: rao.TcpConfig{
				ConnectType:         tcpDetailInfo.TcpConfig.ConnectType,
				IsAutoSend:          tcpDetailInfo.TcpConfig.IsAutoSend,
				ConnectDurationTime: tcpDetailInfo.TcpConfig.ConnectDurationTime,
				SendMsgDurationTime: tcpDetailInfo.TcpConfig.SendMsgDurationTime,
				ConnectTimeoutTime:  tcpDetailInfo.TcpConfig.ConnectTimeoutTime,
				RetryNum:            tcpDetailInfo.TcpConfig.RetryNum,
				RetryInterval:       tcpDetailInfo.TcpConfig.RetryInterval,
			},
		},
		GlobalVariable: globalVariable,
	}
}

func GetSendWebsocketParam(target *model.Target, wsDetailInfo *mao.WebsocketDetail, globalVariable rao.GlobalVariable) rao.RunTargetParam {
	wsHeader := make([]rao.WsQuery, 0, len(wsDetailInfo.WsHeader))
	for _, v := range wsDetailInfo.WsHeader {
		temp := rao.WsQuery{
			IsChecked: v.IsChecked,
			Var:       v.Var,
			Val:       v.Val,
		}
		wsHeader = append(wsHeader, temp)
	}

	wsParam := make([]rao.WsQuery, 0, len(wsDetailInfo.WsParam))
	for _, v := range wsDetailInfo.WsParam {
		temp := rao.WsQuery{
			IsChecked: v.IsChecked,
			Var:       v.Var,
			Val:       v.Val,
		}
		wsParam = append(wsParam, temp)
	}

	wsEvent := make([]rao.WsQuery, 0, len(wsDetailInfo.WsEvent))
	for _, v := range wsDetailInfo.WsEvent {
		temp := rao.WsQuery{
			IsChecked: v.IsChecked,
			Var:       v.Var,
			Val:       v.Val,
		}
		wsEvent = append(wsEvent, temp)
	}

	url := wsDetailInfo.Url
	if wsDetailInfo.EnvInfo.EnvID != 0 {
		url = wsDetailInfo.EnvInfo.PreUrl + url
	}

	return rao.RunTargetParam{
		TargetID:   target.TargetID,
		Name:       target.Name,
		TeamID:     target.TeamID,
		TargetType: target.TargetType,
		WebsocketDetail: rao.WebsocketDetail{
			Url:         url,
			SendMessage: wsDetailInfo.SendMessage,
			MessageType: wsDetailInfo.MessageType,
			WsHeader:    wsHeader,
			WsParam:     wsParam,
			WsEvent:     wsEvent,
			WsConfig: rao.WsConfig{
				ConnectType:         wsDetailInfo.WsConfig.ConnectType,
				IsAutoSend:          wsDetailInfo.WsConfig.IsAutoSend,
				ConnectDurationTime: wsDetailInfo.WsConfig.ConnectDurationTime,
				SendMsgDurationTime: wsDetailInfo.WsConfig.SendMsgDurationTime,
				ConnectTimeoutTime:  wsDetailInfo.WsConfig.ConnectTimeoutTime,
				RetryNum:            wsDetailInfo.WsConfig.RetryNum,
				RetryInterval:       wsDetailInfo.WsConfig.RetryInterval,
			},
		},
		GlobalVariable: globalVariable,
	}
}

func GetSendDubboParam(target *model.Target, dubboDetailInfo *mao.DubboDetail, globalVariable rao.GlobalVariable) rao.RunTargetParam {
	dubboParam := make([]rao.DubboParam, 0, len(dubboDetailInfo.DubboParam))
	for _, v := range dubboDetailInfo.DubboParam {
		temp := rao.DubboParam{
			IsChecked: v.IsChecked,
			ParamType: v.ParamType,
			Var:       v.Var,
			Val:       v.Val,
		}
		dubboParam = append(dubboParam, temp)
	}

	dubboAssert := make([]rao.DubboAssert, 0, len(dubboDetailInfo.DubboAssert))
	for _, v := range dubboDetailInfo.DubboAssert {
		temp := rao.DubboAssert{
			IsChecked:    v.IsChecked,
			ResponseType: v.ResponseType,
			Var:          v.Var,
			Compare:      v.Compare,
			Val:          v.Val,
			Index:        v.Index,
		}
		dubboAssert = append(dubboAssert, temp)
	}

	dubboRegex := make([]rao.DubboRegex, 0, len(dubboDetailInfo.DubboRegex))
	for _, v := range dubboDetailInfo.DubboRegex {
		temp := rao.DubboRegex{
			IsChecked: v.IsChecked,
			Type:      v.Type,
			Var:       v.Var,
			Express:   v.Express,
			Val:       v.Val,
			Index:     v.Index,
		}
		dubboRegex = append(dubboRegex, temp)
	}

	return rao.RunTargetParam{
		TargetID:   target.TargetID,
		Name:       target.Name,
		TeamID:     target.TeamID,
		TargetType: target.TargetType,
		DubboDetail: rao.DubboDetail{
			DubboProtocol: dubboDetailInfo.DubboProtocol,
			ApiName:       dubboDetailInfo.ApiName,
			FunctionName:  dubboDetailInfo.FunctionName,
			DubboParam:    dubboParam,
			DubboAssert:   dubboAssert,
			DubboRegex:    dubboRegex,
			DubboConfig: rao.DubboConfig{
				RegistrationCenterName:    dubboDetailInfo.DubboConfig.RegistrationCenterName,
				RegistrationCenterAddress: dubboDetailInfo.DubboConfig.RegistrationCenterAddress,
				Version:                   dubboDetailInfo.DubboConfig.Version,
			},
		},
		GlobalVariable: globalVariable,
	}
}

func GetSendMqttParam(target *model.Target, mqttDetailInfo *mao.MqttDetail, globalVariable rao.GlobalVariable) rao.RunMqttParam {
	return rao.RunMqttParam{
		TargetID:   target.TargetID,
		Name:       target.Name,
		TeamID:     target.TeamID,
		TargetType: target.TargetType,
		MQTTConfig: rao.MqttDetail{
			Topic:       mqttDetailInfo.Topic,
			SendMessage: mqttDetailInfo.SendMessage,
			CommonConfig: rao.CommonConfig{
				ClientName: mqttDetailInfo.CommonConfig.ClientName,
				Username:   mqttDetailInfo.CommonConfig.Username,
				Password:   mqttDetailInfo.CommonConfig.Password,
				IsEncrypt:  mqttDetailInfo.CommonConfig.IsEncrypt,
				AuthFile: rao.AuthFile{
					FileName: mqttDetailInfo.CommonConfig.AuthFile.FileName,
					FileUrl:  mqttDetailInfo.CommonConfig.AuthFile.FileUrl,
				},
			},
			HigherConfig: rao.HigherConfig{
				ConnectTimeoutTime:  mqttDetailInfo.HigherConfig.ConnectTimeoutTime,
				KeepAliveTime:       mqttDetailInfo.HigherConfig.KeepAliveTime,
				IsAutoRetry:         mqttDetailInfo.HigherConfig.IsAutoRetry,
				RetryNum:            mqttDetailInfo.HigherConfig.RetryNum,
				RetryInterval:       mqttDetailInfo.HigherConfig.RetryInterval,
				MqttVersion:         mqttDetailInfo.HigherConfig.MqttVersion,
				DialogueTimeout:     mqttDetailInfo.HigherConfig.DialogueTimeout,
				IsSaveMessage:       mqttDetailInfo.HigherConfig.IsSaveMessage,
				ServiceQuality:      mqttDetailInfo.HigherConfig.ServiceQuality,
				SendMsgIntervalTime: mqttDetailInfo.HigherConfig.SendMsgIntervalTime,
			},
			WillConfig: rao.WillConfig{
				WillTopic:      mqttDetailInfo.WillConfig.WillTopic,
				IsOpenWill:     mqttDetailInfo.WillConfig.IsOpenWill,
				ServiceQuality: mqttDetailInfo.WillConfig.ServiceQuality,
			},
		},
		GlobalVariable: globalVariable,
	}
}

func GetRunTargetParam(target *model.Target, globalVariable rao.GlobalVariable, api *mao.API) rao.RunTargetParam {
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
