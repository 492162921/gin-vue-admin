package inspection

import (
	"context"
	"fmt"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	inspModel "github.com/flipped-aurora/gin-vue-admin/server/model/inspection"
	inspRequest "github.com/flipped-aurora/gin-vue-admin/server/model/inspection/request"
	"go.uber.org/zap"
)

type ReportService struct{}

func (s *ReportService) Generate(ctx context.Context, inspectionID uint) (*inspModel.InspReport, error) {
	inspection, err := s.inspectionWithDetails(ctx, inspectionID)
	if err != nil {
		return nil, err
	}
	template := renderReportTemplate(inspection)
	content, source := template, "template"
	config := global.GVA_CONFIG.Inspection
	if strings.TrimSpace(config.AIBaseURL) != "" && strings.TrimSpace(config.AIAPIKey) != "" {
		aiContent, aiErr := newOpenAIChatClient(config.AIBaseURL, config.AIAPIKey, config.AIModel).
			Complete(ctx, "你是 Kubernetes 巡检分析助手。请用中文输出简洁、可执行的 Markdown 巡检报告。", template)
		if aiErr == nil {
			content, source = aiContent, "ai"
		} else if global.GVA_LOG != nil {
			global.GVA_LOG.Warn("生成 AI 巡检报告失败，回退到模板", zap.Error(aiErr))
		}
	}
	report := &inspModel.InspReport{
		InspectionID: inspection.ID,
		Title:        fmt.Sprintf("巡检报告 #%d", inspection.ID),
		ContentMD:    content,
		Digest:       reportDigest(inspection.Summary),
		Source:       source,
	}
	if err := global.GVA_DB.WithContext(ctx).Create(report).Error; err != nil {
		return nil, err
	}
	return report, nil
}

func (s *ReportService) List(ctx context.Context, search inspRequest.ReportSearch) ([]inspModel.InspReport, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&inspModel.InspReport{})
	if search.Title != "" {
		db = db.Where("title LIKE ?", "%"+search.Title+"%")
	}
	if search.InspectionID != 0 {
		db = db.Where("inspection_id = ?", search.InspectionID)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := search.LimitOffset()
	var reports []inspModel.InspReport
	if err := db.Order("id DESC").Limit(limit).Offset(offset).Find(&reports).Error; err != nil {
		return nil, 0, err
	}
	return reports, total, nil
}

func (s *ReportService) Get(ctx context.Context, reportID uint) (*inspModel.InspReport, error) {
	var report inspModel.InspReport
	if err := global.GVA_DB.WithContext(ctx).First(&report, reportID).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

func (s *ReportService) Push(ctx context.Context, reportID uint) error {
	report, err := s.Get(ctx, reportID)
	if err != nil {
		return err
	}
	return NotifyReportWebhook(global.GVA_CONFIG.Inspection.WebhookURL, report.Title, truncateRunes(report.Digest, 200))
}

func (s *ReportService) Delete(ctx context.Context, reportID uint) error {
	return global.GVA_DB.WithContext(ctx).Delete(&inspModel.InspReport{}, "id = ?", reportID).Error
}

func (s *ReportService) inspectionWithDetails(ctx context.Context, inspectionID uint) (*inspModel.InspInspection, error) {
	var inspection inspModel.InspInspection
	if err := global.GVA_DB.WithContext(ctx).Preload("Details").First(&inspection, inspectionID).Error; err != nil {
		return nil, err
	}
	return &inspection, nil
}

func renderReportTemplate(inspection *inspModel.InspInspection) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "# 巡检报告 #%d\n\n", inspection.ID)
	fmt.Fprintf(&builder, "- 状态：%s\n- 巡检摘要：%s\n- 异常数量：%d\n\n", inspection.Status, inspection.Summary, inspection.AnomalyCount)
	builder.WriteString("## 异常明细\n\n")
	found := false
	for _, detail := range inspection.Details {
		if detail.IsAnomaly {
			found = true
			fmt.Fprintf(&builder, "- **%s**（%s）：%s\n", detail.ResourceName, detail.RuleType, detail.Message)
		}
	}
	if !found {
		builder.WriteString("- 未发现异常。\n")
	}
	return builder.String()
}

func reportDigest(summary string) string {
	summary = strings.TrimSpace(summary)
	if summary == "" {
		summary = "巡检报告已生成"
	}
	return truncateRunes(summary, 200)
}

func truncateRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}
