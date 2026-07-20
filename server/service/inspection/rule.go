package inspection

import (
	"context"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	inspModel "github.com/flipped-aurora/gin-vue-admin/server/model/inspection"
	inspRequest "github.com/flipped-aurora/gin-vue-admin/server/model/inspection/request"
)

type RuleService struct{}

func (s *RuleService) Create(ctx context.Context, input inspRequest.CreateRule) (*inspModel.InspRule, error) {
	rule := inspModel.InspRule{
		Name:      input.Name,
		RuleType:  input.RuleType,
		Threshold: input.Threshold,
		Scope:     input.Scope,
		Namespace: input.Namespace,
		Enabled:   true,
	}
	if rule.Scope == "" {
		rule.Scope = "cluster"
	}
	if err := global.GVA_DB.WithContext(ctx).Create(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (s *RuleService) Update(ctx context.Context, id uint, input inspRequest.UpdateRule) (*inspModel.InspRule, error) {
	var rule inspModel.InspRule
	if err := global.GVA_DB.WithContext(ctx).First(&rule, id).Error; err != nil {
		return nil, err
	}
	rule.Name = input.Name
	rule.RuleType = input.RuleType
	rule.Threshold = input.Threshold
	rule.Scope = input.Scope
	rule.Namespace = input.Namespace
	if rule.Scope == "" {
		rule.Scope = "cluster"
	}
	if input.Enabled != nil {
		rule.Enabled = *input.Enabled
	}
	if err := global.GVA_DB.WithContext(ctx).Save(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (s *RuleService) Delete(ctx context.Context, id uint) error {
	var count int64
	if err := global.GVA_DB.WithContext(ctx).Model(&inspModel.InspTaskRule{}).Where("insp_rule_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("规则已绑定巡检任务，不能删除")
	}
	return global.GVA_DB.WithContext(ctx).Delete(&inspModel.InspRule{}, id).Error
}

func (s *RuleService) SetEnabled(ctx context.Context, id uint, enabled bool) error {
	return global.GVA_DB.WithContext(ctx).Model(&inspModel.InspRule{}).Where("id = ?", id).Update("enabled", enabled).Error
}

func (s *RuleService) List(ctx context.Context, search inspRequest.RuleSearch) ([]inspModel.InspRule, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&inspModel.InspRule{})
	if search.Name != "" {
		db = db.Where("name LIKE ?", "%"+search.Name+"%")
	}
	if search.RuleType != "" {
		db = db.Where("rule_type = ?", search.RuleType)
	}
	if search.Enabled != nil {
		db = db.Where("enabled = ?", *search.Enabled)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit, offset := search.LimitOffset()
	var rules []inspModel.InspRule
	if err := db.Order("id DESC").Limit(limit).Offset(offset).Find(&rules).Error; err != nil {
		return nil, 0, err
	}
	return rules, total, nil
}
