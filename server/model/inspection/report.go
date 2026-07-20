package inspection

import "github.com/flipped-aurora/gin-vue-admin/server/global"

type InspReport struct {
	global.GVA_MODEL
	InspectionID uint   `json:"inspectionId" gorm:"comment:巡检ID"`
	Title        string `json:"title" gorm:"size:256;comment:标题"`
	ContentMD    string `json:"contentMd" gorm:"type:text;comment:Markdown内容"`
	Digest       string `json:"digest" gorm:"size:512;comment:摘要"`
	Source       string `json:"source" gorm:"size:16;default:template;comment:来源ai/template"`
}

func (InspReport) TableName() string { return "insp_reports" }
