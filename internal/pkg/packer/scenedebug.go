package packer

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
)

func TransMaoSceneDebugsToRaoSceneDebugs(ms []*mao.SceneDebug) []*rao.SceneDebug {
	ret := make([]*rao.SceneDebug, 0)

	for _, m := range ms {
		assertArr := make([]rao.AssertionMsg, 0, len(m.Assert.AssertionMsgs))
		for _, a := range m.Assert.AssertionMsgs {
			assertArr = append(assertArr, rao.AssertionMsg{
				Type:      a.Type,
				Code:      a.Code,
				IsSucceed: a.IsSucceed,
				Msg:       a.Msg,
			})
		}

		regexArr := make([]rao.Reg, 0, len(m.Regex.Regs))
		for _, regexInfo := range m.Regex.Regs {
			regexArr = append(regexArr, rao.Reg{
				Key:   regexInfo.Key,
				Value: regexInfo.Value,
			})
		}

		ret = append(ret, &rao.SceneDebug{
			ApiID:   m.ApiID,
			APIName: m.APIName,
			EventID: m.EventID,
			Assert: rao.AssertObj{
				AssertionMsgs: assertArr,
			},
			Regex: rao.RegexObj{
				Regs: regexArr,
			},
			NextList:       m.NextList,
			RequestBody:    m.RequestBody,
			RequestCode:    m.RequestCode,
			RequestHeader:  m.RequestHeader,
			ResponseBody:   m.ResponseBody,
			ResponseBytes:  m.ResponseBytes,
			ResponseHeader: m.ResponseHeader,
			Status:         m.Status,
			UUID:           m.UUID,
			ResponseTime:   m.ResponseTime,
			RequestType:    m.RequestType,
		})
	}

	return ret
}
