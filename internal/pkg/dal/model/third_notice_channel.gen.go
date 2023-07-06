// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameThirdNoticeChannel = "third_notice_channel"

// ThirdNoticeChannel mapped from table <third_notice_channel>
type ThirdNoticeChannel struct {
	ID        int64          `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"` // 主键id
	Name      string         `gorm:"column:name;not null" json:"name"`                  // 名称
	Type      int32          `gorm:"column:type;not null" json:"type"`                  // 类型 1:飞书  2:企业微信  3:邮箱  4:钉钉
	CreatedAt time.Time      `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
}

// TableName ThirdNoticeChannel's table name
func (*ThirdNoticeChannel) TableName() string {
	return TableNameThirdNoticeChannel
}