// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameTarget = "target"

// Target mapped from table <target>
type Target struct {
	ID                 int64          `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"` // id
	TeamID             int64          `gorm:"column:team_id;not null" json:"team_id"`            // 团队id
	TargetType         string         `gorm:"column:target_type;not null" json:"target_type"`    // 类型：文件夹，接口，分组，场景
	Name               string         `gorm:"column:name;not null" json:"name"`                  // 名称
	ParentID           int64          `gorm:"column:parent_id;not null" json:"parent_id"`        // 父级ID
	Method             string         `gorm:"column:method;not null" json:"method"`              // 方法
	Sort               int32          `gorm:"column:sort;not null" json:"sort"`                  // 排序
	TypeSort           int32          `gorm:"column:type_sort;not null" json:"type_sort"`        // 类型排序
	Status             int32          `gorm:"column:status;not null" json:"status"`
	Version            int32          `gorm:"column:version;not null" json:"version"`
	CreateUserIdentify string         `gorm:"column:create_user_identify;not null" json:"create_user_identify"`
	RecentUserIdentify string         `gorm:"column:recent_user_identify;not null" json:"recent_user_identify"`
	CreatedUserID      int64          `gorm:"column:created_user_id;not null" json:"created_user_id"`
	RecentUserID       int64          `gorm:"column:recent_user_id;not null" json:"recent_user_id"`
	Description        string         `gorm:"column:description" json:"description"`
	Source             int32          `gorm:"column:source;not null" json:"source"`   // 来源1正常，2计划
	PlanID             int64          `gorm:"column:plan_id;not null" json:"plan_id"` // 计划id
	CreatedAt          time.Time      `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`
}

// TableName Target's table name
func (*Target) TableName() string {
	return TableNameTarget
}