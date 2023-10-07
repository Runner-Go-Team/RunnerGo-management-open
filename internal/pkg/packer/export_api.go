package packer

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/log"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"github.com/go-omnibus/proof"
	"go.mongodb.org/mongo-driver/bson"
)

func ExportTransTargetsToRaoAPIDetails(targets []*model.Target, apiMap map[string]*mao.API,
	sqlMap map[string]*mao.SqlDetailForMg, tcpMap map[string]*mao.TcpDetail,
	websocketMap map[string]*mao.WebsocketDetail, dubboMap map[string]*mao.DubboDetail,
	folderMap map[string]*model.Target) []map[string]interface{} {
	ret := make([]map[string]interface{}, 0, len(targets))

	globalVariable := rao.GlobalVariable{}

	for _, target := range targets {
		if api, ok := apiMap[target.TargetID]; ok {
			ret = append(ret, ExportTransTargetToRaoAPIDetail(target, api, globalVariable))
		}
		if sql, ok := sqlMap[target.TargetID]; ok {
			ret = append(ret, ExportTransTargetToRaoSqlDetail(target, sql, globalVariable))
		}

		if tcp, ok := tcpMap[target.TargetID]; ok {
			ret = append(ret, ExportTransTargetToRaoTcpDetail(target, tcp, globalVariable))
		}
		if websocket, ok := websocketMap[target.TargetID]; ok {
			ret = append(ret, ExportTransTargetToRaoWebsocketDetail(target, websocket, globalVariable))
		}
		//if mqtt, ok := mqttMap[target.TargetID]; ok {
		//	ret = append(ret, TransTargetToRaoMqttDetail(target, mqtt, globalVariable))
		//}
		if dubbo, ok := dubboMap[target.TargetID]; ok {
			ret = append(ret, ExportTransTargetToRaoDubboDetail(target, dubbo, globalVariable))
		}

		if folder, ok := folderMap[target.TargetID]; ok {
			ret = append(ret, ExportTransTargetToRaoFolderDetail(folder))
		}
	}
	return ret
}

func ExportTransTargetToRaoAPIDetail(target *model.Target, api *mao.API, globalVariable rao.GlobalVariable) map[string]interface{} {
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

	res := make(map[string]interface{}, 30)
	res["target_id"] = target.TargetID
	res["parent_id"] = target.ParentID
	res["target_type"] = target.TargetType
	res["team_id"] = target.TeamID
	res["name"] = target.Name
	res["method"] = target.Method
	res["sort"] = target.Sort
	res["type_sort"] = target.TypeSort
	res["version"] = target.Version
	res["description"] = target.Description
	res["created_time_sec"] = target.CreatedAt.Unix()
	res["updated_time_sec"] = target.UpdatedAt.Unix()
	res["global_variable"] = globalVariable
	res["url"] = api.URL
	res["request"] = rao.Request{
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
	}
	res["env_info"] = rao.EnvInfo{
		EnvID:       api.EnvInfo.EnvID,
		EnvName:     api.EnvInfo.EnvName,
		ServiceID:   api.EnvInfo.ServiceID,
		ServiceName: api.EnvInfo.ServiceName,
		PreUrl:      api.EnvInfo.PreUrl,
	}
	return res
}

func ExportTransTargetToRaoSqlDetail(target *model.Target, sql *mao.SqlDetailForMg, globalVariable rao.GlobalVariable) map[string]interface{} {
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

	res := make(map[string]interface{}, 30)
	res["target_id"] = target.TargetID
	res["parent_id"] = target.ParentID
	res["target_type"] = target.TargetType
	res["team_id"] = target.TeamID
	res["name"] = target.Name
	res["method"] = target.Method
	res["sort"] = target.Sort
	res["type_sort"] = target.TypeSort
	res["version"] = target.Version
	res["description"] = target.Description
	res["created_time_sec"] = target.CreatedAt.Unix()
	res["updated_time_sec"] = target.UpdatedAt.Unix()
	res["global_variable"] = globalVariable
	res["env_info"] = rao.EnvInfo{
		EnvID:       sql.EnvInfo.EnvID,
		EnvName:     sql.EnvInfo.EnvName,
		ServiceID:   sql.EnvInfo.ServiceID,
		ServiceName: sql.EnvInfo.ServiceName,
		PreUrl:      sql.EnvInfo.PreUrl,
		DatabaseID:  sql.EnvInfo.DatabaseID,
		ServerName:  sql.EnvInfo.ServerName,
	}
	res["sql_detail"] = rao.SqlDetail{
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
	}
	return res
}

func ExportTransTargetToRaoTcpDetail(target *model.Target, tcp *mao.TcpDetail, globalVariable rao.GlobalVariable) map[string]interface{} {
	res := make(map[string]interface{}, 30)
	res["target_id"] = target.TargetID
	res["parent_id"] = target.ParentID
	res["target_type"] = target.TargetType
	res["team_id"] = target.TeamID
	res["name"] = target.Name
	res["method"] = target.Method
	res["sort"] = target.Sort
	res["type_sort"] = target.TypeSort
	res["version"] = target.Version
	res["description"] = target.Description
	res["created_time_sec"] = target.CreatedAt.Unix()
	res["updated_time_sec"] = target.UpdatedAt.Unix()
	res["global_variable"] = globalVariable
	res["env_info"] = rao.EnvInfo{
		EnvID:       tcp.EnvInfo.EnvID,
		EnvName:     tcp.EnvInfo.EnvName,
		ServiceID:   tcp.EnvInfo.ServiceID,
		ServiceName: tcp.EnvInfo.ServiceName,
		PreUrl:      tcp.EnvInfo.PreUrl,
	}
	res["tcp_detail"] = rao.TcpDetail{
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
	}
	return res
}

func ExportTransTargetToRaoWebsocketDetail(target *model.Target, websocket *mao.WebsocketDetail, globalVariable rao.GlobalVariable) map[string]interface{} {
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

	res := make(map[string]interface{}, 30)
	res["target_id"] = target.TargetID
	res["parent_id"] = target.ParentID
	res["target_type"] = target.TargetType
	res["team_id"] = target.TeamID
	res["name"] = target.Name
	res["method"] = target.Method
	res["sort"] = target.Sort
	res["type_sort"] = target.TypeSort
	res["version"] = target.Version
	res["description"] = target.Description
	res["created_time_sec"] = target.CreatedAt.Unix()
	res["updated_time_sec"] = target.UpdatedAt.Unix()
	res["global_variable"] = globalVariable
	res["env_info"] = rao.EnvInfo{
		EnvID:       websocket.EnvInfo.EnvID,
		EnvName:     websocket.EnvInfo.EnvName,
		ServiceID:   websocket.EnvInfo.ServiceID,
		ServiceName: websocket.EnvInfo.ServiceName,
		PreUrl:      websocket.EnvInfo.PreUrl,
	}
	res["websocket_detail"] = rao.WebsocketDetail{
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
	}
	return res
}

func ExportTransTargetToRaoDubboDetail(target *model.Target, dubbo *mao.DubboDetail, globalVariable rao.GlobalVariable) map[string]interface{} {
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

	res := make(map[string]interface{}, 30)
	res["target_id"] = target.TargetID
	res["parent_id"] = target.ParentID
	res["target_type"] = target.TargetType
	res["team_id"] = target.TeamID
	res["name"] = target.Name
	res["method"] = target.Method
	res["sort"] = target.Sort
	res["type_sort"] = target.TypeSort
	res["version"] = target.Version
	res["description"] = target.Description
	res["created_time_sec"] = target.CreatedAt.Unix()
	res["updated_time_sec"] = target.UpdatedAt.Unix()
	res["global_variable"] = globalVariable
	res["env_info"] = rao.EnvInfo{
		EnvID:       dubbo.EnvInfo.EnvID,
		EnvName:     dubbo.EnvInfo.EnvName,
		ServiceID:   dubbo.EnvInfo.ServiceID,
		ServiceName: dubbo.EnvInfo.ServiceName,
		PreUrl:      dubbo.EnvInfo.PreUrl,
	}
	res["dubbo_detail"] = rao.DubboDetail{
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
	}
	return res
}

func ExportTransTargetToRaoFolderDetail(folder *model.Target) map[string]interface{} {
	res := make(map[string]interface{}, 30)
	res["target_id"] = folder.TargetID
	res["parent_id"] = folder.ParentID
	res["target_type"] = folder.TargetType
	res["team_id"] = folder.TeamID
	res["name"] = folder.Name
	res["sort"] = folder.Sort
	res["type_sort"] = folder.TypeSort
	res["version"] = folder.Version
	res["description"] = folder.Description
	res["created_time_sec"] = folder.CreatedAt.Unix()
	res["updated_time_sec"] = folder.UpdatedAt.Unix()
	return res
}
