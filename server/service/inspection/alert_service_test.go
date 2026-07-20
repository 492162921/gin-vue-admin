package inspection

import (
	"context"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	commonRequest "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	inspModel "github.com/flipped-aurora/gin-vue-admin/server/model/inspection"
	inspRequest "github.com/flipped-aurora/gin-vue-admin/server/model/inspection/request"
)

func TestAlertServiceFiltersAndUpdatesStatus(t *testing.T) {
	testutil.NewMemoryDB(t, &inspModel.InspAlert{})
	ctx := context.Background()
	alerts := []inspModel.InspAlert{
		{InspectionID: 1, RuleType: "node_cpu", Level: "warning", Status: "open", Content: "cpu", CollectedAt: time.Now()},
		{InspectionID: 1, RuleType: "pod_failed", Level: "critical", Status: "open", Content: "pod", CollectedAt: time.Now()},
	}
	if err := defaultAlertService.Create(ctx, alerts); err != nil {
		t.Fatalf("create alerts: %v", err)
	}

	list, total, err := defaultAlertService.List(ctx, inspRequest.AlertSearch{
		PageInfo: commonRequest.PageInfo{Page: 1, PageSize: 10},
		Level:    "critical",
	})
	if err != nil {
		t.Fatalf("list alerts: %v", err)
	}
	if total != 1 || len(list) != 1 || list[0].RuleType != "pod_failed" {
		t.Fatalf("unexpected filtered alerts: total=%d list=%+v", total, list)
	}
	if err := defaultAlertService.UpdateStatus(ctx, list[0].ID, "acknowledged"); err != nil {
		t.Fatalf("update status: %v", err)
	}
	updated, _, err := defaultAlertService.List(ctx, inspRequest.AlertSearch{
		PageInfo: commonRequest.PageInfo{Page: 1, PageSize: 10},
		Status:   "acknowledged",
	})
	if err != nil {
		t.Fatalf("list updated alerts: %v", err)
	}
	if len(updated) != 1 || updated[0].ID != list[0].ID {
		t.Fatalf("unexpected acknowledged alerts: %+v", updated)
	}
}
