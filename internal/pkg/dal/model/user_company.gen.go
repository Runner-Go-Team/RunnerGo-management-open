// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameUserCompany = "user_company"

// UserCompany mapped from table <user_company>
type UserCompany struct {
	ID           int64          `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`              // 主键id
	UserID       string         `gorm:"column:user_id;not null" json:"user_id"`                         // 用户id
	CompanyID    string         `gorm:"column:company_id;not null" json:"company_id"`                   // 企业id
	InviteUserID string         `gorm:"column:invite_user_id;not null;default:0" json:"invite_user_id"` // 邀请人id
	InviteTime   time.Time      `gorm:"column:invite_time" json:"invite_time"`                          // 邀请时间
	Status       int32          `gorm:"column:status;not null;default:1" json:"status"`                 // 状态：1-正常，2-已禁用
	CreatedAt    time.Time      `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
}

// TableName UserCompany's table name
func (*UserCompany) TableName() string {
	return TableNameUserCompany
}