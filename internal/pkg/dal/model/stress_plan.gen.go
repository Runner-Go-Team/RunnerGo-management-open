// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameStressPlan = "stress_plan"

// StressPlan mapped from table <stress_plan>
type StressPlan struct {
	ID           int64          `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`                      // 主键ID
	PlanID       string         `gorm:"column:plan_id;not null;default:0" json:"plan_id"`                       // 计划ID
	TeamID       string         `gorm:"column:team_id;not null" json:"team_id"`                                 // 团队ID
	RankID       int64          `gorm:"column:rank_id;not null" json:"rank_id"`                                 // 序号ID
	PlanName     string         `gorm:"column:plan_name;not null" json:"plan_name"`                             // 计划名称
	TaskType     int32          `gorm:"column:task_type;not null" json:"task_type"`                             // 计划类型：1-普通任务，2-定时任务
	TaskMode     int32          `gorm:"column:task_mode;not null" json:"task_mode"`                             // 压测类型: 1-并发模式，2-阶梯模式，3-错误率模式，4-响应时间模式，5-每秒请求数模式，6-每秒事务数模式
	Status       int32          `gorm:"column:status;not null;default:1" json:"status"`                         // 计划状态1:未开始,2:进行中
	CreateUserID string         `gorm:"column:create_user_id;not null;default:0" json:"create_user_id"`         // 创建人id
	RunUserID    string         `gorm:"column:run_user_id;not null;default:0" json:"run_user_id"`               // 运行人id
	Remark       string         `gorm:"column:remark" json:"remark"`                                            // 备注
	RunCount     int64          `gorm:"column:run_count" json:"run_count"`                                      // 运行次数
	CreatedAt    time.Time      `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"` // 创建时间
	UpdatedAt    time.Time      `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"` // 修改时间
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`                                    // 删除时间
}

// TableName StressPlan's table name
func (*StressPlan) TableName() string {
	return TableNameStressPlan
}