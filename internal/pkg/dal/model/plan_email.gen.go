// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNamePlanEmail = "plan_email"

// PlanEmail mapped from table <plan_email>
type PlanEmail struct {
	ID        int64          `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`                      // 主键
	PlanID    int64          `gorm:"column:plan_id;not null" json:"plan_id"`                                 // 计划ID
	TeamID    int64          `gorm:"column:team_id;not null" json:"team_id"`                                 // 团队ID
	Email     string         `gorm:"column:email" json:"email"`                                              // 邮箱
	CreatedAt time.Time      `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"` // 创建时间
	UpdatedAt time.Time      `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"` // 修改时间
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`                                    // 删除时间
}

// TableName PlanEmail's table name
func (*PlanEmail) TableName() string {
	return TableNamePlanEmail
}
