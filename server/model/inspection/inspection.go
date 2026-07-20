package inspection

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

type InspInspection struct {
	global.GVA_MODEL
	TaskID       uint       `json:"taskId" gorm:"comment:任务ID"`
	Task         *InspTask  `json:"task,omitempty" gorm:"foreignKey:TaskID"`
	Status       string     `json:"status" gorm:"size:16;default:running;comment:状态"`
	AnomalyCount int        `json:"anomalyCount" gorm:"default:0;comment:异常数"`
	Summary      string     `json:"summary" gorm:"type:text;comment:摘要"`
	StartedAt    time.Time  `json:"startedAt" gorm:"comment:开始时间"`
	FinishedAt   *time.Time `json:"finishedAt,omitempty" gorm:"comment:结束时间"`
	Details      []InspInspectionDetail `json:"details,omitempty" gorm:"foreignKey:InspectionID"`
}

func (InspInspection) TableName() string { return "insp_inspections" }

type InspInspectionDetail struct {
	global.GVA_MODEL
	InspectionID uint    `json:"inspectionId" gorm:"comment:巡检ID"`
	RuleType     string  `json:"ruleType" gorm:"size:64;comment:规则类型"`
	ResourceType string  `json:"resourceType" gorm:"size:64;comment:资源类型"`
	ResourceName string  `json:"resourceName" gorm:"size:256;comment:资源名"`
	Message      string  `json:"message" gorm:"type:text;comment:消息"`
	IsAnomaly    bool    `json:"isAnomaly" gorm:"default:false;comment:是否异常"`
	MetricValue  float64 `json:"metricValue" gorm:"comment:指标值"`
}

func (InspInspectionDetail) TableName() string { return "insp_inspection_details" }
