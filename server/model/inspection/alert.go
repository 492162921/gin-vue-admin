package inspection

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

type InspAlert struct {
	global.GVA_MODEL
	InspectionID uint      `json:"inspectionId" gorm:"comment:巡检ID"`
	RuleType     string    `json:"ruleType" gorm:"size:64;comment:规则类型"`
	Level        string    `json:"level" gorm:"size:16;default:warning;comment:级别"`
	Status       string    `json:"status" gorm:"size:16;default:open;comment:状态open/acknowledged/closed"`
	Content      string    `json:"content" gorm:"type:text;comment:内容"`
	CollectedAt  time.Time `json:"collectedAt" gorm:"comment:采集时间"`
}

func (InspAlert) TableName() string { return "insp_alerts" }
