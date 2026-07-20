package inspection

import (
	"context"
	"strings"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	inspModel "github.com/flipped-aurora/gin-vue-admin/server/model/inspection"
)

func TestDeleteRuleBlockedWhenBoundToTask(t *testing.T) {
	db := testutil.NewMemoryDB(t, &inspModel.InspRule{}, &inspModel.InspTask{}, &inspModel.InspCluster{})

	rule := inspModel.InspRule{Name: "节点 CPU", RuleType: "node_cpu", Enabled: true}
	if err := db.Create(&rule).Error; err != nil {
		t.Fatalf("创建规则失败: %v", err)
	}
	task := inspModel.InspTask{Name: "每日巡检", ClusterID: 1, Rules: []inspModel.InspRule{rule}}
	if err := db.Create(&task).Error; err != nil {
		t.Fatalf("创建任务失败: %v", err)
	}

	err := new(RuleService).Delete(context.Background(), rule.ID)
	if err == nil {
		t.Fatal("已绑定任务的规则删除应失败")
	}
	if !strings.Contains(err.Error(), "任务") {
		t.Fatalf("错误信息应说明规则仍被任务引用，实际为: %v", err)
	}
}
