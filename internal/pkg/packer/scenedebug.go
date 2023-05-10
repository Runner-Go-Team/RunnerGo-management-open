package packer

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
)

func TransMaoSceneDebugsToRaoSceneDebugs(ms []*mao.SceneDebug) []*rao.SceneDebug {
	ret := make([]*rao.SceneDebug, 0)

	for _, m := range ms {

		var as []*rao.DebugAssertion
		for _, a := range m.Assertion {
			as = append(as, &rao.DebugAssertion{
				Code:      a.Code,
				IsSucceed: a.IsSucceed,
				Msg:       a.Msg,
			})
		}

		ret = append(ret, &rao.SceneDebug{
			ApiID:          m.ApiID,
			APIName:        m.APIName,
			Assertion:      as,
			EventID:        m.EventID,
			NextList:       m.NextList,
			Regex:          m.Regex,
			RequestBody:    m.RequestBody,
			RequestCode:    m.RequestCode,
			RequestHeader:  m.RequestHeader,
			ResponseBody:   m.ResponseBody,
			ResponseBytes:  m.ResponseBytes,
			ResponseHeader: m.ResponseHeader,
			Status:         m.Status,
			UUID:           m.UUID,
			ResponseTime:   m.ResponseTime,
		})
	}

	return ret
}
