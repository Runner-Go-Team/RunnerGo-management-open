// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameUser = "user"

// User mapped from table <user>
type User struct {
	ID           int64          `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	UserID       string         `gorm:"column:user_id;not null" json:"user_id"`               // 用户id
	Account      string         `gorm:"column:account;not null" json:"account"`               // 账号
	Email        string         `gorm:"column:email;not null" json:"email"`                   // 邮箱
	Mobile       string         `gorm:"column:mobile;not null" json:"mobile"`                 // 手机号
	Password     string         `gorm:"column:password;not null" json:"password"`             // 密码
	Nickname     string         `gorm:"column:nickname;not null" json:"nickname"`             // 昵称
	Avatar       string         `gorm:"column:avatar" json:"avatar"`                          // 头像
	WechatOpenID string         `gorm:"column:wechat_open_id;not null" json:"wechat_open_id"` // 微信开放的唯一id
	UtmSource    string         `gorm:"column:utm_source;not null" json:"utm_source"`         // 渠道来源
	LastLoginAt  time.Time      `gorm:"column:last_login_at" json:"last_login_at"`            // 最近登录时间
	CreatedAt    time.Time      `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
}

// TableName User's table name
func (*User) TableName() string {
	return TableNameUser
}
