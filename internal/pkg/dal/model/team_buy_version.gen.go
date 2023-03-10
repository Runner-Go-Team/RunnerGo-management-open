// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameTeamBuyVersion = "team_buy_version"

// TeamBuyVersion mapped from table <team_buy_version>
type TeamBuyVersion struct {
	ID               int32          `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`                      // 主键id
	Title            string         `gorm:"column:title;not null" json:"title"`                                     // 购买套餐名称
	UnitPrice        float64        `gorm:"column:unit_price;not null" json:"unit_price"`                           // 单人单月定价
	UnitPriceExplain string         `gorm:"column:unit_price_explain;not null" json:"unit_price_explain"`           // 单人单月定价说明
	Detail           string         `gorm:"column:detail;not null" json:"detail"`                                   // 套餐详情
	MinUserNum       int64          `gorm:"column:min_user_num;not null" json:"min_user_num"`                       // 最少团队成员数
	MaxUserNum       int64          `gorm:"column:max_user_num;not null" json:"max_user_num"`                       // 最大团队成员数
	MaxConcurrence   int64          `gorm:"column:max_concurrence;not null" json:"max_concurrence"`                 // 最大并发数
	MaxAPINum        int64          `gorm:"column:max_api_num;not null" json:"max_api_num"`                         // 最大接口数
	MaxRunTime       int64          `gorm:"column:max_run_time;not null" json:"max_run_time"`                       // 最大运行时长
	GiveVunNum       int64          `gorm:"column:give_vun_num;not null" json:"give_vun_num"`                       // 赠送VUM配额
	CreatedAt        time.Time      `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"` // 创建时间
	UpdatedAt        time.Time      `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"` // 修改时间
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at"`                                    // 删除时间
}

// TableName TeamBuyVersion's table name
func (*TeamBuyVersion) TableName() string {
	return TableNameTeamBuyVersion
}
