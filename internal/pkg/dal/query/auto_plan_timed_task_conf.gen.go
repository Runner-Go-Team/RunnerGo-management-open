// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package query

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"gorm.io/plugin/dbresolver"

	"github.com/Runner-Go-Team/RunnerGo-management-open/internal/pkg/dal/model"
)

func newAutoPlanTimedTaskConf(db *gorm.DB, opts ...gen.DOOption) autoPlanTimedTaskConf {
	_autoPlanTimedTaskConf := autoPlanTimedTaskConf{}

	_autoPlanTimedTaskConf.autoPlanTimedTaskConfDo.UseDB(db, opts...)
	_autoPlanTimedTaskConf.autoPlanTimedTaskConfDo.UseModel(&model.AutoPlanTimedTaskConf{})

	tableName := _autoPlanTimedTaskConf.autoPlanTimedTaskConfDo.TableName()
	_autoPlanTimedTaskConf.ALL = field.NewAsterisk(tableName)
	_autoPlanTimedTaskConf.ID = field.NewInt32(tableName, "id")
	_autoPlanTimedTaskConf.PlanID = field.NewString(tableName, "plan_id")
	_autoPlanTimedTaskConf.TeamID = field.NewString(tableName, "team_id")
	_autoPlanTimedTaskConf.Frequency = field.NewInt32(tableName, "frequency")
	_autoPlanTimedTaskConf.TaskExecTime = field.NewInt64(tableName, "task_exec_time")
	_autoPlanTimedTaskConf.TaskCloseTime = field.NewInt64(tableName, "task_close_time")
	_autoPlanTimedTaskConf.FixedIntervalStartTime = field.NewInt64(tableName, "fixed_interval_start_time")
	_autoPlanTimedTaskConf.FixedIntervalTime = field.NewInt32(tableName, "fixed_interval_time")
	_autoPlanTimedTaskConf.FixedRunNum = field.NewInt32(tableName, "fixed_run_num")
	_autoPlanTimedTaskConf.FixedIntervalTimeType = field.NewInt32(tableName, "fixed_interval_time_type")
	_autoPlanTimedTaskConf.TaskType = field.NewInt32(tableName, "task_type")
	_autoPlanTimedTaskConf.TaskMode = field.NewInt32(tableName, "task_mode")
	_autoPlanTimedTaskConf.SceneRunOrder = field.NewInt32(tableName, "scene_run_order")
	_autoPlanTimedTaskConf.TestCaseRunOrder = field.NewInt32(tableName, "test_case_run_order")
	_autoPlanTimedTaskConf.Status = field.NewInt32(tableName, "status")
	_autoPlanTimedTaskConf.RunUserID = field.NewString(tableName, "run_user_id")
	_autoPlanTimedTaskConf.CreatedAt = field.NewTime(tableName, "created_at")
	_autoPlanTimedTaskConf.UpdatedAt = field.NewTime(tableName, "updated_at")
	_autoPlanTimedTaskConf.DeletedAt = field.NewField(tableName, "deleted_at")

	_autoPlanTimedTaskConf.fillFieldMap()

	return _autoPlanTimedTaskConf
}

type autoPlanTimedTaskConf struct {
	autoPlanTimedTaskConfDo autoPlanTimedTaskConfDo

	ALL                    field.Asterisk
	ID                     field.Int32  // 表id
	PlanID                 field.String // 计划id
	TeamID                 field.String // 团队id
	Frequency              field.Int32  // 任务执行频次: 0-一次，1-每天，2-每周，3-每月，4-固定时间间隔
	TaskExecTime           field.Int64  // 任务执行时间
	TaskCloseTime          field.Int64  // 任务结束时间
	FixedIntervalStartTime field.Int64  // 固定时间间隔开始时间
	FixedIntervalTime      field.Int32  // 固定间隔时间
	FixedRunNum            field.Int32  // 固定执行次数
	FixedIntervalTimeType  field.Int32  // 固定间隔时间类型：0-分钟，1-小时
	TaskType               field.Int32  // 任务类型：1-普通任务，2-定时任务
	TaskMode               field.Int32  // 运行模式：1-按照用例执行
	SceneRunOrder          field.Int32  // 场景运行次序：1-顺序执行，2-同时执行
	TestCaseRunOrder       field.Int32  // 测试用例运行次序：1-顺序执行，2-同时执行
	Status                 field.Int32  // 任务状态：0-未启用，1-运行中，2-已过期
	RunUserID              field.String // 运行人用户ID
	CreatedAt              field.Time   // 创建时间
	UpdatedAt              field.Time   // 更新时间
	DeletedAt              field.Field  // 删除时间

	fieldMap map[string]field.Expr
}

func (a autoPlanTimedTaskConf) Table(newTableName string) *autoPlanTimedTaskConf {
	a.autoPlanTimedTaskConfDo.UseTable(newTableName)
	return a.updateTableName(newTableName)
}

func (a autoPlanTimedTaskConf) As(alias string) *autoPlanTimedTaskConf {
	a.autoPlanTimedTaskConfDo.DO = *(a.autoPlanTimedTaskConfDo.As(alias).(*gen.DO))
	return a.updateTableName(alias)
}

func (a *autoPlanTimedTaskConf) updateTableName(table string) *autoPlanTimedTaskConf {
	a.ALL = field.NewAsterisk(table)
	a.ID = field.NewInt32(table, "id")
	a.PlanID = field.NewString(table, "plan_id")
	a.TeamID = field.NewString(table, "team_id")
	a.Frequency = field.NewInt32(table, "frequency")
	a.TaskExecTime = field.NewInt64(table, "task_exec_time")
	a.TaskCloseTime = field.NewInt64(table, "task_close_time")
	a.FixedIntervalStartTime = field.NewInt64(table, "fixed_interval_start_time")
	a.FixedIntervalTime = field.NewInt32(table, "fixed_interval_time")
	a.FixedRunNum = field.NewInt32(table, "fixed_run_num")
	a.FixedIntervalTimeType = field.NewInt32(table, "fixed_interval_time_type")
	a.TaskType = field.NewInt32(table, "task_type")
	a.TaskMode = field.NewInt32(table, "task_mode")
	a.SceneRunOrder = field.NewInt32(table, "scene_run_order")
	a.TestCaseRunOrder = field.NewInt32(table, "test_case_run_order")
	a.Status = field.NewInt32(table, "status")
	a.RunUserID = field.NewString(table, "run_user_id")
	a.CreatedAt = field.NewTime(table, "created_at")
	a.UpdatedAt = field.NewTime(table, "updated_at")
	a.DeletedAt = field.NewField(table, "deleted_at")

	a.fillFieldMap()

	return a
}

func (a *autoPlanTimedTaskConf) WithContext(ctx context.Context) *autoPlanTimedTaskConfDo {
	return a.autoPlanTimedTaskConfDo.WithContext(ctx)
}

func (a autoPlanTimedTaskConf) TableName() string { return a.autoPlanTimedTaskConfDo.TableName() }

func (a autoPlanTimedTaskConf) Alias() string { return a.autoPlanTimedTaskConfDo.Alias() }

func (a *autoPlanTimedTaskConf) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := a.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (a *autoPlanTimedTaskConf) fillFieldMap() {
	a.fieldMap = make(map[string]field.Expr, 19)
	a.fieldMap["id"] = a.ID
	a.fieldMap["plan_id"] = a.PlanID
	a.fieldMap["team_id"] = a.TeamID
	a.fieldMap["frequency"] = a.Frequency
	a.fieldMap["task_exec_time"] = a.TaskExecTime
	a.fieldMap["task_close_time"] = a.TaskCloseTime
	a.fieldMap["fixed_interval_start_time"] = a.FixedIntervalStartTime
	a.fieldMap["fixed_interval_time"] = a.FixedIntervalTime
	a.fieldMap["fixed_run_num"] = a.FixedRunNum
	a.fieldMap["fixed_interval_time_type"] = a.FixedIntervalTimeType
	a.fieldMap["task_type"] = a.TaskType
	a.fieldMap["task_mode"] = a.TaskMode
	a.fieldMap["scene_run_order"] = a.SceneRunOrder
	a.fieldMap["test_case_run_order"] = a.TestCaseRunOrder
	a.fieldMap["status"] = a.Status
	a.fieldMap["run_user_id"] = a.RunUserID
	a.fieldMap["created_at"] = a.CreatedAt
	a.fieldMap["updated_at"] = a.UpdatedAt
	a.fieldMap["deleted_at"] = a.DeletedAt
}

func (a autoPlanTimedTaskConf) clone(db *gorm.DB) autoPlanTimedTaskConf {
	a.autoPlanTimedTaskConfDo.ReplaceConnPool(db.Statement.ConnPool)
	return a
}

func (a autoPlanTimedTaskConf) replaceDB(db *gorm.DB) autoPlanTimedTaskConf {
	a.autoPlanTimedTaskConfDo.ReplaceDB(db)
	return a
}

type autoPlanTimedTaskConfDo struct{ gen.DO }

func (a autoPlanTimedTaskConfDo) Debug() *autoPlanTimedTaskConfDo {
	return a.withDO(a.DO.Debug())
}

func (a autoPlanTimedTaskConfDo) WithContext(ctx context.Context) *autoPlanTimedTaskConfDo {
	return a.withDO(a.DO.WithContext(ctx))
}

func (a autoPlanTimedTaskConfDo) ReadDB() *autoPlanTimedTaskConfDo {
	return a.Clauses(dbresolver.Read)
}

func (a autoPlanTimedTaskConfDo) WriteDB() *autoPlanTimedTaskConfDo {
	return a.Clauses(dbresolver.Write)
}

func (a autoPlanTimedTaskConfDo) Session(config *gorm.Session) *autoPlanTimedTaskConfDo {
	return a.withDO(a.DO.Session(config))
}

func (a autoPlanTimedTaskConfDo) Clauses(conds ...clause.Expression) *autoPlanTimedTaskConfDo {
	return a.withDO(a.DO.Clauses(conds...))
}

func (a autoPlanTimedTaskConfDo) Returning(value interface{}, columns ...string) *autoPlanTimedTaskConfDo {
	return a.withDO(a.DO.Returning(value, columns...))
}

func (a autoPlanTimedTaskConfDo) Not(conds ...gen.Condition) *autoPlanTimedTaskConfDo {
	return a.withDO(a.DO.Not(conds...))
}

func (a autoPlanTimedTaskConfDo) Or(conds ...gen.Condition) *autoPlanTimedTaskConfDo {
	return a.withDO(a.DO.Or(conds...))
}

func (a autoPlanTimedTaskConfDo) Select(conds ...field.Expr) *autoPlanTimedTaskConfDo {
	return a.withDO(a.DO.Select(conds...))
}

func (a autoPlanTimedTaskConfDo) Where(conds ...gen.Condition) *autoPlanTimedTaskConfDo {
	return a.withDO(a.DO.Where(conds...))
}

func (a autoPlanTimedTaskConfDo) Exists(subquery interface{ UnderlyingDB() *gorm.DB }) *autoPlanTimedTaskConfDo {
	return a.Where(field.CompareSubQuery(field.ExistsOp, nil, subquery.UnderlyingDB()))
}

func (a autoPlanTimedTaskConfDo) Order(conds ...field.Expr) *autoPlanTimedTaskConfDo {
	return a.withDO(a.DO.Order(conds...))
}

func (a autoPlanTimedTaskConfDo) Distinct(cols ...field.Expr) *autoPlanTimedTaskConfDo {
	return a.withDO(a.DO.Distinct(cols...))
}

func (a autoPlanTimedTaskConfDo) Omit(cols ...field.Expr) *autoPlanTimedTaskConfDo {
	return a.withDO(a.DO.Omit(cols...))
}

func (a autoPlanTimedTaskConfDo) Join(table schema.Tabler, on ...field.Expr) *autoPlanTimedTaskConfDo {
	return a.withDO(a.DO.Join(table, on...))
}

func (a autoPlanTimedTaskConfDo) LeftJoin(table schema.Tabler, on ...field.Expr) *autoPlanTimedTaskConfDo {
	return a.withDO(a.DO.LeftJoin(table, on...))
}

func (a autoPlanTimedTaskConfDo) RightJoin(table schema.Tabler, on ...field.Expr) *autoPlanTimedTaskConfDo {
	return a.withDO(a.DO.RightJoin(table, on...))
}

func (a autoPlanTimedTaskConfDo) Group(cols ...field.Expr) *autoPlanTimedTaskConfDo {
	return a.withDO(a.DO.Group(cols...))
}

func (a autoPlanTimedTaskConfDo) Having(conds ...gen.Condition) *autoPlanTimedTaskConfDo {
	return a.withDO(a.DO.Having(conds...))
}

func (a autoPlanTimedTaskConfDo) Limit(limit int) *autoPlanTimedTaskConfDo {
	return a.withDO(a.DO.Limit(limit))
}

func (a autoPlanTimedTaskConfDo) Offset(offset int) *autoPlanTimedTaskConfDo {
	return a.withDO(a.DO.Offset(offset))
}

func (a autoPlanTimedTaskConfDo) Scopes(funcs ...func(gen.Dao) gen.Dao) *autoPlanTimedTaskConfDo {
	return a.withDO(a.DO.Scopes(funcs...))
}

func (a autoPlanTimedTaskConfDo) Unscoped() *autoPlanTimedTaskConfDo {
	return a.withDO(a.DO.Unscoped())
}

func (a autoPlanTimedTaskConfDo) Create(values ...*model.AutoPlanTimedTaskConf) error {
	if len(values) == 0 {
		return nil
	}
	return a.DO.Create(values)
}

func (a autoPlanTimedTaskConfDo) CreateInBatches(values []*model.AutoPlanTimedTaskConf, batchSize int) error {
	return a.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (a autoPlanTimedTaskConfDo) Save(values ...*model.AutoPlanTimedTaskConf) error {
	if len(values) == 0 {
		return nil
	}
	return a.DO.Save(values)
}

func (a autoPlanTimedTaskConfDo) First() (*model.AutoPlanTimedTaskConf, error) {
	if result, err := a.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*model.AutoPlanTimedTaskConf), nil
	}
}

func (a autoPlanTimedTaskConfDo) Take() (*model.AutoPlanTimedTaskConf, error) {
	if result, err := a.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*model.AutoPlanTimedTaskConf), nil
	}
}

func (a autoPlanTimedTaskConfDo) Last() (*model.AutoPlanTimedTaskConf, error) {
	if result, err := a.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*model.AutoPlanTimedTaskConf), nil
	}
}

func (a autoPlanTimedTaskConfDo) Find() ([]*model.AutoPlanTimedTaskConf, error) {
	result, err := a.DO.Find()
	return result.([]*model.AutoPlanTimedTaskConf), err
}

func (a autoPlanTimedTaskConfDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.AutoPlanTimedTaskConf, err error) {
	buf := make([]*model.AutoPlanTimedTaskConf, 0, batchSize)
	err = a.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (a autoPlanTimedTaskConfDo) FindInBatches(result *[]*model.AutoPlanTimedTaskConf, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return a.DO.FindInBatches(result, batchSize, fc)
}

func (a autoPlanTimedTaskConfDo) Attrs(attrs ...field.AssignExpr) *autoPlanTimedTaskConfDo {
	return a.withDO(a.DO.Attrs(attrs...))
}

func (a autoPlanTimedTaskConfDo) Assign(attrs ...field.AssignExpr) *autoPlanTimedTaskConfDo {
	return a.withDO(a.DO.Assign(attrs...))
}

func (a autoPlanTimedTaskConfDo) Joins(fields ...field.RelationField) *autoPlanTimedTaskConfDo {
	for _, _f := range fields {
		a = *a.withDO(a.DO.Joins(_f))
	}
	return &a
}

func (a autoPlanTimedTaskConfDo) Preload(fields ...field.RelationField) *autoPlanTimedTaskConfDo {
	for _, _f := range fields {
		a = *a.withDO(a.DO.Preload(_f))
	}
	return &a
}

func (a autoPlanTimedTaskConfDo) FirstOrInit() (*model.AutoPlanTimedTaskConf, error) {
	if result, err := a.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*model.AutoPlanTimedTaskConf), nil
	}
}

func (a autoPlanTimedTaskConfDo) FirstOrCreate() (*model.AutoPlanTimedTaskConf, error) {
	if result, err := a.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*model.AutoPlanTimedTaskConf), nil
	}
}

func (a autoPlanTimedTaskConfDo) FindByPage(offset int, limit int) (result []*model.AutoPlanTimedTaskConf, count int64, err error) {
	result, err = a.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = a.Offset(-1).Limit(-1).Count()
	return
}

func (a autoPlanTimedTaskConfDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = a.Count()
	if err != nil {
		return
	}

	err = a.Offset(offset).Limit(limit).Scan(result)
	return
}

func (a autoPlanTimedTaskConfDo) Scan(result interface{}) (err error) {
	return a.DO.Scan(result)
}

func (a autoPlanTimedTaskConfDo) Delete(models ...*model.AutoPlanTimedTaskConf) (result gen.ResultInfo, err error) {
	return a.DO.Delete(models)
}

func (a *autoPlanTimedTaskConfDo) withDO(do gen.Dao) *autoPlanTimedTaskConfDo {
	a.DO = *do.(*gen.DO)
	return a
}
