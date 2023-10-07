package mao

import (
	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/rao"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

type SceneOperator struct {
	SceneID       string   `bson:"scene_id"`
	OperatorID    string   `bson:"operator_id"`
	ActionDetail  bson.Raw `bson:"action_detail"`
	Settings      bson.Raw `bson:"settings"`
	Asserts       bson.Raw `bson:"asserts"`
	DataWithdraws bson.Raw `bson:"data_withdraws"`
}

type Settings struct {
	Settings *rao.AutomationSettings `bson:"settings"`
}

type Asserts struct {
	Asserts []*rao.AutomationAssert `bson:"asserts"`
}

type DataWithdraws struct {
	DataWithdraws []*rao.DataWithdraw `bson:"data_withdraws"`
}

type ActionDetail struct {
	Detail *rao.ActionDetail `bson:"detail"`
}

type UISendScene struct {
	SceneID            string                 `json:"scene_id" bson:"scene_id"`
	Name               string                 `json:"name" bson:"name"`
	OperatorTotalNum   int                    `json:"operator_total_num" bson:"operator_total_num"`
	OperatorSuccessNum int                    `json:"operator_success_num" bson:"operator_success_num"`
	OperatorErrorNum   int                    `json:"operator_error_num" bson:"operator_error_num"`
	OperatorUnExecNum  int                    `json:"operator_un_exec_num" bson:"operator_un_exec_num"`
	AssertTotalNum     int                    `json:"assert_total_num" bson:"assert_total_num"`
	Operators          []*UISendSceneOperator `json:"operators" bson:"operators"`
}

type UISendSceneOperator struct {
	ReportId        string    `json:"report_id" bson:"report_id"`
	TeamID          string    `json:"team_id" bson:"team_id"`
	SceneID         string    `json:"scene_id" bson:"scene_id"`
	SceneName       string    `json:"scene_name" bson:"scene_name"`
	OperatorID      string    `json:"operator_id" bson:"operator_id"`
	ParentID        string    `json:"parent_id" bson:"parent_id"`
	Name            string    `json:"name" bson:"name"`
	Sort            int32     `json:"sort" bson:"sort"`
	Type            string    `json:"type" bson:"type"`
	Action          string    `json:"action" bson:"action"`
	RunStatus       int32     `json:"run_status" bson:"run_status"`       // 1:未测 2:成功  3:失败
	ExecTime        float64   `json:"exec_time" bson:"exec_time"`         //  运行时长
	RunEndTimes     int64     `json:"run_end_times" bson:"run_end_times"` // 运行结束时间
	Status          string    `json:"status" bson:"status"`               // 状态
	Msg             string    `json:"msg" bson:"msg"`
	Screenshot      string    `json:"screenshot" bson:"screenshot"`
	ScreenshotUrl   string    `json:"screenshot_url" bson:"screenshot_url"`
	End             bool      `json:"end" bson:"end"`
	IsMulti         bool      `json:"is_multi" bson:"is_multi"` // 是否展示多条
	AssertTotalNum  int       `json:"assert_total_num" bson:"assert_total_num"`
	CreatedAt       time.Time `json:"created_at" bson:"created_at"`
	Detail          bson.Raw  `json:"detail" bson:"detail"`
	AssertResults   bson.Raw  `json:"assert_results" bson:"assert_results"`
	MultiResults    bson.Raw  `json:"multi_results" bson:"multi_results"` // 多条数据结果
	WithdrawResults bson.Raw  `json:"withdraw_results" bson:"withdraw_results"`
}

type UIOperatorDetail struct {
	Operator *SceneOperator `json:"operator" bson:"operator"`
}

type AssertResults struct {
	Asserts []*rao.UIEngineAssertion `json:"asserts" bson:"asserts"`
}

type WithdrawResults struct {
	Withdraws []*rao.UIEngineDataWithdraw `json:"withdraws" bson:"withdraws"`
}

type MultiResults struct {
	MultiResults []*rao.UIEngineResultDataMsg `json:"multi_results"`
}
