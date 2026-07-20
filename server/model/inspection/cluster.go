package inspection

import "github.com/flipped-aurora/gin-vue-admin/server/global"

type InspCluster struct {
	global.GVA_MODEL
	Name           string `json:"name" gorm:"size:128;comment:集群名"`
	KubeconfigPath string `json:"-" gorm:"size:512;comment:kubeconfig路径"`
	Status         string `json:"status" gorm:"size:32;default:unknown"`
	K8sVersion     string `json:"k8sVersion" gorm:"size:64"`
	NodeCount      int    `json:"nodeCount" gorm:"default:0"`
}

func (InspCluster) TableName() string { return "insp_clusters" }
