package response

import (
	"net/http"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/biz/errno"

	"github.com/gin-gonic/gin"
)

type response struct {
	Code int         `json:"code"`
	Em   string      `json:"em,omitempty"`
	Et   string      `json:"et,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

func display(c *gin.Context, code int, em string, et string, data interface{}) {
	respData := response{
		Code: code,
		Em:   em,
		Et:   et,
		Data: data,
	}
	c.JSON(http.StatusOK, respData)
}

// ErrorWithMsg 返回错误 附带更多信息
func ErrorWithMsg(c *gin.Context, code int, msg string) {
	if m, ok := errno.CodeAlertMap[code]; ok {
		msg = m + " " + msg
	}
	display(c, code, errno.CodeMsgMap[code], msg, struct{}{})
}

// ErrorWithMsgAndData  返回错误 附带更多信息,同时带着data返回值
func ErrorWithMsgAndData(c *gin.Context, code int, msg string, data interface{}) {
	if m, ok := errno.CodeAlertMap[code]; ok {
		msg = m + " " + msg
	}
	display(c, code, errno.CodeMsgMap[code], msg, data)
}

// SuccessWithData 返回成功并携带数据
func SuccessWithData(c *gin.Context, data interface{}) {
	display(c, errno.Ok, errno.CodeMsgMap[errno.Ok], errno.CodeAlertMap[errno.Ok], data)
}

func Success(c *gin.Context) {
	display(c, errno.Ok, errno.CodeMsgMap[errno.Ok], errno.CodeAlertMap[errno.Ok], struct{}{})
}
