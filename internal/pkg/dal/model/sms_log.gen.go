// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameSmsLog = "sms_log"

// SmsLog mapped from table <sms_log>
type SmsLog struct {
	ID                       int64          `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`                                                        // 主键id
	Type                     int32          `gorm:"column:type;not null" json:"type"`                                                                         // 短信类型: 1-注册，2-登录，3-找回密码
	Mobile                   string         `gorm:"column:mobile;not null" json:"mobile"`                                                                     // 手机号
	Content                  string         `gorm:"column:content;not null" json:"content"`                                                                   // 短信内容
	VerifyCode               string         `gorm:"column:verify_code;not null" json:"verify_code"`                                                           // 验证码
	VerifyCodeExpirationTime time.Time      `gorm:"column:verify_code_expiration_time;not null;default:CURRENT_TIMESTAMP" json:"verify_code_expiration_time"` // 验证码有效时间
	ClientIP                 string         `gorm:"column:client_ip;not null" json:"client_ip"`                                                               // 客户端IP
	SendStatus               int32          `gorm:"column:send_status;not null;default:1" json:"send_status"`                                                 // 发送状态：1-成功 2-失败
	VerifyStatus             int32          `gorm:"column:verify_status;not null;default:1" json:"verify_status"`                                             // 校验状态：1-未校验 2-已校验
	SendResponse             string         `gorm:"column:send_response;not null" json:"send_response"`                                                       // 短信服务响应
	CreatedAt                time.Time      `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`                                   // 创建时间
	UpdatedAt                time.Time      `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`                                   // 修改时间
	DeletedAt                gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`                                                                      // 删除时间
}

// TableName SmsLog's table name
func (*SmsLog) TableName() string {
	return TableNameSmsLog
}
