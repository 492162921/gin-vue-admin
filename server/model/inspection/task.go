package inspection

import "github.com/flipped-aurora/gin-vue-admin/server/global"

type InspTask struct {
	global.GVA_MODEL
	Name         string     `json:"name" gorm:"size:128;comment:任务名"`
	ClusterID    uint       `json:"clusterId" gorm:"comment:集群ID"`
	Cluster      *InspCluster `json:"cluster,omitempty" gorm:"foreignKey:ClusterID"`
	ScheduleType string     `json:"scheduleType" gorm:"size:16;default:cron;comment:调度类型"`
	CronExpr     string     `json:"cronExpr" gorm:"size:64;comment:cron表达式"`
	Status       string     `json:"status" gorm:"size:16;default:active;comment:状态"`
	TimedTaskID  *uint      `json:"timedTaskId" gorm:"comment:关联sys_timed_tasks"`
	Rules        []InspRule `json:"rules,omitempty" gorm:"many2many:insp_task_rules;"`
}

func (InspTask) TableName() string { return "insp_tasks" }

type InspTaskRule struct {
	TaskID uint `gorm:"primaryKey;comment:任务ID"`
	RuleID uint `gorm:"primaryKey;comment:规则ID"`
}

func (InspTaskRule) TableName() string { return "insp_task_rules" }
