package inspection

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	inspModel "github.com/flipped-aurora/gin-vue-admin/server/model/inspection"
	inspRequest "github.com/flipped-aurora/gin-vue-admin/server/model/inspection/request"
	sysModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"gorm.io/datatypes"
)

var defaultEngine = NewEngine(nil)

type TaskService struct{}

func (s *TaskService) Create(ctx context.Context, input inspRequest.CreateTask) (*inspModel.InspTask, error) {
	rules, err := findTaskRules(ctx, input.RuleIDs)
	if err != nil {
		return nil, err
	}
	task := inspModel.InspTask{
		Name: input.Name, ClusterID: input.ClusterID, CronExpr: input.CronExpr,
		ScheduleType: "cron", Status: "active", Rules: rules,
	}
	if err := global.GVA_DB.WithContext(ctx).Create(&task).Error; err != nil {
		return nil, err
	}
	if err := s.syncTimedTask(ctx, &task); err != nil {
		return nil, err
	}
	return s.Get(ctx, task.ID)
}

func (s *TaskService) Update(ctx context.Context, id uint, input inspRequest.UpdateTask) (*inspModel.InspTask, error) {
	task, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	rules, err := findTaskRules(ctx, input.RuleIDs)
	if err != nil {
		return nil, err
	}
	task.Name = input.Name
	task.ClusterID = input.ClusterID
	task.CronExpr = input.CronExpr
	if input.Status != "" {
		task.Status = input.Status
	}
	if err := global.GVA_DB.WithContext(ctx).Model(task).Association("Rules").Replace(rules); err != nil {
		return nil, err
	}
	if err := global.GVA_DB.WithContext(ctx).Save(task).Error; err != nil {
		return nil, err
	}
	if err := s.syncTimedTask(ctx, task); err != nil {
		return nil, err
	}
	return s.Get(ctx, task.ID)
}

func (s *TaskService) Delete(ctx context.Context, id uint) error {
	task, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := global.GVA_DB.WithContext(ctx).Model(task).Association("Rules").Clear(); err != nil {
		return err
	}
	if task.TimedTaskID != nil {
		if err := system.TimedTaskServiceApp.DeleteTimedTask(ctx, *task.TimedTaskID); err != nil {
			return err
		}
	}
	return global.GVA_DB.WithContext(ctx).Delete(task).Error
}

func (s *TaskService) Get(ctx context.Context, id uint) (*inspModel.InspTask, error) {
	var task inspModel.InspTask
	if err := global.GVA_DB.WithContext(ctx).Preload("Cluster").Preload("Rules").First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *TaskService) List(ctx context.Context, search inspRequest.TaskSearch) ([]inspModel.InspTask, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&inspModel.InspTask{})
	if search.Name != "" {
		db = db.Where("name LIKE ?", "%"+search.Name+"%")
	}
	if search.ClusterID != 0 {
		db = db.Where("cluster_id = ?", search.ClusterID)
	}
	if search.Status != "" {
		db = db.Where("status = ?", search.Status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := search.LimitOffset()
	var tasks []inspModel.InspTask
	if err := db.Preload("Cluster").Preload("Rules").Order("id DESC").Limit(limit).Offset(offset).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

func (s *TaskService) RunNow(ctx context.Context, taskID uint) (*inspModel.InspInspection, error) {
	return defaultEngine.Run(ctx, taskID)
}

func (s *TaskService) ListInspections(ctx context.Context, taskID uint, info request.PageInfo) ([]inspModel.InspInspection, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&inspModel.InspInspection{}).Where("task_id = ?", taskID)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := info.LimitOffset()
	var inspections []inspModel.InspInspection
	if err := db.Preload("Details").Order("id DESC").Limit(limit).Offset(offset).Find(&inspections).Error; err != nil {
		return nil, 0, err
	}
	return inspections, total, nil
}

func findTaskRules(ctx context.Context, ids []uint) ([]inspModel.InspRule, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("至少选择一条巡检规则")
	}
	var rules []inspModel.InspRule
	if err := global.GVA_DB.WithContext(ctx).Where("id IN ?", ids).Find(&rules).Error; err != nil {
		return nil, err
	}
	if len(rules) != len(ids) {
		return nil, fmt.Errorf("包含不存在的巡检规则")
	}
	return rules, nil
}

func (s *TaskService) syncTimedTask(ctx context.Context, inspectionTask *inspModel.InspTask) error {
	if inspectionTask.CronExpr == "" {
		if inspectionTask.TimedTaskID != nil {
			return system.TimedTaskServiceApp.ToggleTimedTask(ctx, *inspectionTask.TimedTaskID, false)
		}
		return nil
	}

	params, err := json.Marshal(struct {
		TaskID uint `json:"taskId"`
	}{TaskID: inspectionTask.ID})
	if err != nil {
		return err
	}
	timedTask := sysModel.SysTimedTask{
		Name:         fmt.Sprintf("insp-task-%d", inspectionTask.ID),
		Description:  "执行 K8s 巡检任务",
		Spec:         inspectionTask.CronExpr,
		WithSeconds:  false,
		ExecutorType: sysModel.TimedTaskExecutorMethod,
		MethodName:   "RunInspectionTask",
		Params:       datatypes.JSON(params),
		Enabled:      inspectionTask.Status == "active",
	}
	if inspectionTask.TimedTaskID == nil {
		if err := system.TimedTaskServiceApp.CreateTimedTask(ctx, &timedTask); err != nil {
			return err
		}
		inspectionTask.TimedTaskID = &timedTask.ID
		return global.GVA_DB.WithContext(ctx).Model(inspectionTask).Update("timed_task_id", timedTask.ID).Error
	}
	timedTask.ID = *inspectionTask.TimedTaskID
	return system.TimedTaskServiceApp.UpdateTimedTask(ctx, &timedTask)
}
