package inspection

import (
	"context"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	inspModel "github.com/flipped-aurora/gin-vue-admin/server/model/inspection"
	inspRequest "github.com/flipped-aurora/gin-vue-admin/server/model/inspection/request"
)

var defaultAlertService = AlertService{}

type AlertService struct{}

func (s *AlertService) Create(ctx context.Context, alerts []inspModel.InspAlert) error {
	if len(alerts) == 0 {
		return nil
	}
	return global.GVA_DB.WithContext(ctx).Create(&alerts).Error
}

func (s *AlertService) List(ctx context.Context, search inspRequest.AlertSearch) ([]inspModel.InspAlert, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&inspModel.InspAlert{})
	if search.Status != "" {
		db = db.Where("status = ?", search.Status)
	}
	if search.Level != "" {
		db = db.Where("level = ?", search.Level)
	}
	if search.RuleType != "" {
		db = db.Where("rule_type = ?", search.RuleType)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := search.LimitOffset()
	var alerts []inspModel.InspAlert
	if err := db.Order("id DESC").Limit(limit).Offset(offset).Find(&alerts).Error; err != nil {
		return nil, 0, err
	}
	return alerts, total, nil
}

func (s *AlertService) UpdateStatus(ctx context.Context, id uint, status string) error {
	if !validAlertStatus(status) {
		return fmt.Errorf("无效告警状态: %s", status)
	}
	return global.GVA_DB.WithContext(ctx).Model(&inspModel.InspAlert{}).Where("id = ?", id).Update("status", status).Error
}

func (s *AlertService) Delete(ctx context.Context, id uint) error {
	return global.GVA_DB.WithContext(ctx).Delete(&inspModel.InspAlert{}, "id = ?", id).Error
}

func (s *AlertService) DeleteByIDs(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return global.GVA_DB.WithContext(ctx).Where("id IN ?", ids).Delete(&inspModel.InspAlert{}).Error
}

func validAlertStatus(status string) bool {
	return status == "open" || status == "acknowledged" || status == "closed"
}
