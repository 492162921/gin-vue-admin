package inspection

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	inspModel "github.com/flipped-aurora/gin-vue-admin/server/model/inspection"
)

func TestGenerateFallsBackToTemplate(t *testing.T) {
	testutil.NewMemoryDB(t, &inspModel.InspInspection{}, &inspModel.InspInspectionDetail{}, &inspModel.InspReport{})
	originalConfig := global.GVA_CONFIG.Inspection
	t.Cleanup(func() { global.GVA_CONFIG.Inspection = originalConfig })
	global.GVA_CONFIG.Inspection.AIBaseURL = ""
	global.GVA_CONFIG.Inspection.AIAPIKey = ""

	inspectionID := createReportInspection(t)
	report, err := (&ReportService{}).Generate(context.Background(), inspectionID)
	if err != nil {
		t.Fatalf("generate report: %v", err)
	}
	if report.Source != "template" {
		t.Fatalf("source = %q, want template", report.Source)
	}
	if strings.TrimSpace(report.ContentMD) == "" {
		t.Fatal("template content is empty")
	}
}

func TestGenerateUsesAIWhenConfigured(t *testing.T) {
	testutil.NewMemoryDB(t, &inspModel.InspInspection{}, &inspModel.InspInspectionDetail{}, &inspModel.InspReport{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatal("missing authorization header")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"# AI 巡检报告\nAI 摘要"}}]}`))
	}))
	t.Cleanup(server.Close)
	originalConfig := global.GVA_CONFIG.Inspection
	t.Cleanup(func() { global.GVA_CONFIG.Inspection = originalConfig })
	global.GVA_CONFIG.Inspection.AIBaseURL = server.URL
	global.GVA_CONFIG.Inspection.AIAPIKey = "test-key"
	global.GVA_CONFIG.Inspection.AIModel = "test-model"

	report, err := (&ReportService{}).Generate(context.Background(), createReportInspection(t))
	if err != nil {
		t.Fatalf("generate report: %v", err)
	}
	if report.Source != "ai" {
		t.Fatalf("source = %q, want ai", report.Source)
	}
	if report.ContentMD != "# AI 巡检报告\nAI 摘要" {
		t.Fatalf("unexpected AI content: %q", report.ContentMD)
	}
}

func createReportInspection(t *testing.T) uint {
	t.Helper()
	now := time.Now()
	inspection := inspModel.InspInspection{
		TaskID:       1,
		Status:       "success",
		AnomalyCount: 1,
		Summary:      "巡检完成，共 2 项，异常 1 项",
		StartedAt:    now,
		FinishedAt:   &now,
	}
	if err := global.GVA_DB.Create(&inspection).Error; err != nil {
		t.Fatalf("create inspection: %v", err)
	}
	detail := inspModel.InspInspectionDetail{
		InspectionID: inspection.ID,
		RuleType:     "pod_failed",
		ResourceType: "pod",
		ResourceName: "default/api",
		Message:      "Phase Failed",
		IsAnomaly:    true,
	}
	if err := global.GVA_DB.Create(&detail).Error; err != nil {
		t.Fatalf("create detail: %v", err)
	}
	return inspection.ID
}
