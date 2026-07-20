package inspection

import "github.com/flipped-aurora/gin-vue-admin/server/global"

type InspRule struct {
	global.GVA_MODEL
	Name      string  `json:"name" gorm:"size:128;comment:规则名"`
	RuleType  string  `json:"ruleType" gorm:"size:64;comment:规则类型"`
	Threshold float64 `json:"threshold" gorm:"comment:阈值"`
	Scope     string  `json:"scope" gorm:"size:64;default:cluster;comment:作用域cluster/namespace"`
	Namespace string  `json:"namespace" gorm:"size:128;comment:命名空间"`
	Enabled   bool    `json:"enabled" gorm:"default:true;comment:是否启用"`
}

func (InspRule) TableName() string { return "insp_rules" }
