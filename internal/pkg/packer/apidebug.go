package packer

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/mao"
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
)

func TransMaoAPIDebugToRaoAPIDebug(m *mao.APIDebug) *rao.APIDebug {
	as := make([]rao.AssertionMsg, 0, len(m.Assert.AssertionMsgs))
	for _, a := range m.Assert.AssertionMsgs {
		as = append(as, rao.AssertionMsg{
			Type:      a.Type,
			Code:      a.Code,
			IsSucceed: a.IsSucceed,
			Msg:       a.Msg,
		})
	}

	regexArr := make([]rao.Reg, 0, len(m.Regex.Regs))
	for _, regInfo := range m.Regex.Regs {
		regexArr = append(regexArr, rao.Reg{
			Key:   regInfo.Key,
			Value: regInfo.Value,
		})
	}

	return &rao.APIDebug{
		ApiID:   m.ApiID,
		APIName: m.APIName,
		EventID: m.EventID,
		Assert: rao.AssertObj{
			AssertionMsgs: as,
		},
		Regex: rao.RegexObj{
			Regs: regexArr,
		},
		RequestBody:           m.RequestBody,
		RequestCode:           m.RequestCode,
		RequestHeader:         m.RequestHeader,
		RequestTime:           m.RequestTime,
		ResponseBody:          m.ResponseBody,
		ResponseBytes:         m.ResponseBytes,
		ResponseHeader:        m.ResponseHeader,
		ResponseTime:          m.ResponseTime,
		ResponseLen:           m.ResponseLen,
		ResponseStatusMessage: m.ResponseStatusMessage,
		UUID:                  m.UUID,
		Status:                m.Status,
	}
}
