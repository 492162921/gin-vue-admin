package inspection

import (
	"context"
	"sort"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	inspModel "github.com/flipped-aurora/gin-vue-admin/server/model/inspection"
)

type DashboardService struct{}

type DashboardSummary struct {
	ClusterCount         int64 `json:"clusterCount"`
	NodeCount            int64 `json:"nodeCount"`
	TodayInspectionCount int64 `json:"todayInspectionCount"`
	OpenAlertCount       int64 `json:"openAlertCount"`
}

type DashboardResourceUsage struct {
	ClusterID   uint    `json:"clusterId"`
	ClusterName string  `json:"clusterName"`
	CPUUsage    float64 `json:"cpuUsage"`
	MemoryUsage float64 `json:"memoryUsage"`
}

type DashboardTrendItem struct {
	Date         string `json:"date"`
	AnomalyCount int    `json:"anomalyCount"`
}

type DashboardDistributionItem struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type DashboardAlertDistribution struct {
	ByLevel   []DashboardDistributionItem `json:"byLevel"`
	ByCluster []DashboardDistributionItem `json:"byCluster"`
}

func (s *DashboardService) Summary(ctx context.Context) (DashboardSummary, error) {
	db := global.GVA_DB.WithContext(ctx)
	summary := DashboardSummary{}
	if err := db.Model(&inspModel.InspCluster{}).Count(&summary.ClusterCount).Error; err != nil {
		return summary, err
	}
	if err := db.Model(&inspModel.InspCluster{}).Select("COALESCE(SUM(node_count), 0)").Scan(&summary.NodeCount).Error; err != nil {
		return summary, err
	}
	dayStart := time.Now().Truncate(24 * time.Hour)
	if err := db.Model(&inspModel.InspInspection{}).Where("started_at >= ?", dayStart).Count(&summary.TodayInspectionCount).Error; err != nil {
		return summary, err
	}
	if err := db.Model(&inspModel.InspAlert{}).Where("status = ?", "open").Count(&summary.OpenAlertCount).Error; err != nil {
		return summary, err
	}
	return summary, nil
}

func (s *DashboardService) ResourceUsage(ctx context.Context) ([]DashboardResourceUsage, error) {
	var inspections []inspModel.InspInspection
	if err := global.GVA_DB.WithContext(ctx).
		Preload("Task.Cluster").
		Where("status = ?", "success").
		Order("started_at DESC").
		Find(&inspections).Error; err != nil {
		return nil, err
	}

	result := make([]DashboardResourceUsage, 0)
	seenClusters := make(map[uint]struct{})
	for _, inspection := range inspections {
		if inspection.Task == nil || inspection.Task.Cluster == nil {
			continue
		}
		cluster := inspection.Task.Cluster
		if _, seen := seenClusters[cluster.ID]; seen {
			continue
		}
		seenClusters[cluster.ID] = struct{}{}
		var metrics []inspModel.InspInspectionDetail
		if err := global.GVA_DB.WithContext(ctx).
			Where("inspection_id = ? AND rule_type IN ?", inspection.ID, []string{"node_cpu", "cpu_high", "node_memory", "mem_high"}).
			Find(&metrics).Error; err != nil {
			return nil, err
		}
		cpu, memory, cpuCount, memoryCount := 0.0, 0.0, 0, 0
		for _, metric := range metrics {
			switch normalizeRuleType(metric.RuleType) {
			case "node_cpu":
				cpu += metric.MetricValue
				cpuCount++
			case "node_memory":
				memory += metric.MetricValue
				memoryCount++
			}
		}
		if cpuCount == 0 && memoryCount == 0 {
			continue
		}
		if cpuCount > 0 {
			cpu /= float64(cpuCount)
		}
		if memoryCount > 0 {
			memory /= float64(memoryCount)
		}
		result = append(result, DashboardResourceUsage{ClusterID: cluster.ID, ClusterName: cluster.Name, CPUUsage: cpu, MemoryUsage: memory})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ClusterName < result[j].ClusterName })
	return result, nil
}

func (s *DashboardService) InspectionTrend(ctx context.Context, days int) ([]DashboardTrendItem, error) {
	if days <= 0 {
		days = 7
	}
	if days > 90 {
		days = 90
	}
	today := time.Now().Truncate(24 * time.Hour)
	start := today.AddDate(0, 0, -(days - 1))
	var inspections []inspModel.InspInspection
	if err := global.GVA_DB.WithContext(ctx).
		Where("started_at >= ? AND started_at < ?", start, today.AddDate(0, 0, 1)).
		Find(&inspections).Error; err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	for _, inspection := range inspections {
		counts[inspection.StartedAt.Format("2006-01-02")] += inspection.AnomalyCount
	}
	result := make([]DashboardTrendItem, 0, days)
	for offset := 0; offset < days; offset++ {
		date := start.AddDate(0, 0, offset).Format("2006-01-02")
		result = append(result, DashboardTrendItem{Date: date, AnomalyCount: counts[date]})
	}
	return result, nil
}

func (s *DashboardService) AlertDistribution(ctx context.Context) (DashboardAlertDistribution, error) {
	db := global.GVA_DB.WithContext(ctx)
	result := DashboardAlertDistribution{
		ByLevel:   make([]DashboardDistributionItem, 0),
		ByCluster: make([]DashboardDistributionItem, 0),
	}
	if err := db.Model(&inspModel.InspAlert{}).
		Select("level AS name, COUNT(*) AS count").
		Group("level").
		Order("level ASC").
		Scan(&result.ByLevel).Error; err != nil {
		return result, err
	}
	if err := db.Model(&inspModel.InspAlert{}).
		Select("insp_clusters.name AS name, COUNT(insp_alerts.id) AS count").
		Joins("JOIN insp_inspections ON insp_inspections.id = insp_alerts.inspection_id").
		Joins("JOIN insp_tasks ON insp_tasks.id = insp_inspections.task_id").
		Joins("JOIN insp_clusters ON insp_clusters.id = insp_tasks.cluster_id").
		Group("insp_clusters.id, insp_clusters.name").
		Order("insp_clusters.name ASC").
		Scan(&result.ByCluster).Error; err != nil {
		return result, err
	}
	return result, nil
}
